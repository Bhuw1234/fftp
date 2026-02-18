#!/usr/bin/env python3
"""
DEparrow Bootstrap Server
Meta-OS Control Plane - Replaces default Bacalhau bootstrap

Features:
- Node registration and discovery
- Orchestrator management
- Credit-based job submission
- PicoClaw agent integration
- WebSocket real-time updates
- Tool execution API
"""

import asyncio
import json
import logging
import os
import sys
import time
import uuid
from collections import defaultdict
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Callable, Set
from dataclasses import dataclass, asdict, field
from enum import Enum

import aiohttp
from aiohttp import web, WSMsgType
import jwt
import redis.asyncio as redis
from pydantic import BaseModel, Field, validator, ValidationError
import asyncpg

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration
class Config:
    """DEparrow Bootstrap Configuration"""
    HOST = os.getenv("DEPARROW_HOST", "0.0.0.0")
    PORT = int(os.getenv("DEPARROW_PORT", "8080"))
    JWT_SECRET = os.getenv("DEPARROW_JWT_SECRET", "deparrow-secret-key-change-me")
    JWT_ALGORITHM = "HS256"
    JWT_EXPIRY_HOURS = 24
    
    # Database
    DATABASE_URL = os.getenv("DEPARROW_DATABASE_URL", "postgresql://deparrow:deparrow@localhost/deparrow")
    REDIS_URL = os.getenv("DEPARROW_REDIS_URL", "redis://localhost:6379")
    
    # Credit system
    CREDIT_EARNING_RATE = float(os.getenv("DEPARROW_CREDIT_RATE", "0.1"))  # credits per CPU-hour
    CREDIT_SUBMISSION_COST = float(os.getenv("DEPARROW_SUBMISSION_COST", "1.0"))
    MIN_CREDIT_BALANCE = float(os.getenv("DEPARROW_MIN_BALANCE", "10.0"))
    
    # Network
    ORCHESTRATOR_PORT = 4222
    NODE_REGISTRATION_TTL = 300  # 5 minutes

# Data Models
class NodeStatus(str, Enum):
    ONLINE = "online"
    OFFLINE = "offline"
    MAINTENANCE = "maintenance"
    SUSPENDED = "suspended"

class NodeArchitecture(str, Enum):
    X86_64 = "x86_64"
    ARM64 = "arm64"

class JobStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"

class NodeRegistration(BaseModel):
    """Node registration request"""
    node_id: str
    public_key: str
    resources: Dict[str, Any]
    arch: NodeArchitecture
    labels: Dict[str, str] = Field(default_factory=dict)

class JobSubmission(BaseModel):
    """Job submission request"""
    job_id: str
    user_id: str
    spec: Dict[str, Any]
    credit_cost: float
    orchestrator: str

class CreditTransaction(BaseModel):
    """Credit transaction"""
    transaction_id: str
    user_id: str
    amount: float
    type: str  # "earn", "spend", "transfer"
    description: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)


# ============ PicoClaw Agent Models ============

class AgentStatus(str, Enum):
    """Agent status enumeration"""
    INITIALIZING = "initializing"
    IDLE = "idle"
    WORKING = "working"
    ERROR = "error"
    OFFLINE = "offline"


class AgentConfig(BaseModel):
    """Agent configuration model"""
    model: str = Field(default="claude-3-5-sonnet-20241022", description="LLM model to use")
    workspace: str = Field(default=".", description="Working directory")
    max_iterations: int = Field(default=10, description="Maximum iterations per task")
    auto_approve: bool = Field(default=False, description="Auto-approve tool executions")
    tools_enabled: List[str] = Field(default_factory=lambda: ["job", "credit", "node", "wallet"], description="Enabled tools")
    temperature: float = Field(default=0.7, description="LLM temperature")
    system_prompt: Optional[str] = Field(default=None, description="Custom system prompt")


class AgentRegistration(BaseModel):
    """Agent registration request"""
    agent_id: str = Field(default_factory=lambda: f"agent-{uuid.uuid4().hex[:12]}")
    name: str = Field(..., description="Agent display name")
    node_id: str = Field(..., description="Associated compute node ID")
    config: AgentConfig = Field(default_factory=AgentConfig)


class ToolDefinition(BaseModel):
    """Tool definition for DEparrow tools"""
    name: str
    description: str
    category: str
    parameters: Dict[str, Any]
    requires_auth: bool = True
    rate_limit: int = 60  # requests per minute


class ToolExecutionRequest(BaseModel):
    """Tool execution request"""
    tool_name: str
    parameters: Dict[str, Any]
    agent_id: str
    async_exec: bool = False


class ToolExecutionResult(BaseModel):
    """Tool execution result"""
    execution_id: str
    tool_name: str
    status: str  # "success", "error", "pending"
    result: Optional[Any] = None
    error: Optional[str] = None
    duration_ms: float
    timestamp: datetime = Field(default_factory=datetime.utcnow)

# Database Models
class ContributionTier(str, Enum):
    BRONZE = "bronze"
    SILVER = "silver"
    GOLD = "gold"
    DIAMOND = "diamond"
    LEGENDARY = "legendary"

@dataclass
class Node:
    node_id: str
    public_key: str
    arch: NodeArchitecture
    resources: Dict[str, Any]
    status: NodeStatus
    last_seen: datetime
    credits_earned: float = 0.0
    labels: Dict[str, str] = None
    # Contribution tracking
    cpu_cores: int = 0
    cpu_usage_hours: float = 0.0
    gpu_count: int = 0
    gpu_model: str = ""
    gpu_usage_hours: float = 0.0
    memory_gb: float = 0.0
    live_gflops: float = 0.0
    location: Dict[str, float] = None  # {"lat": 0, "lng": 0}
    
    def __post_init__(self):
        if self.labels is None:
            self.labels = {}
        if self.location is None:
            self.location = {"lat": 0.0, "lng": 0.0}
    
    def get_tier(self) -> ContributionTier:
        """Calculate node tier based on contribution"""
        total_hours = self.cpu_usage_hours + self.gpu_usage_hours
        if total_hours >= 10000:
            return ContributionTier.LEGENDARY
        elif total_hours >= 5000:
            return ContributionTier.DIAMOND
        elif total_hours >= 1000:
            return ContributionTier.GOLD
        elif total_hours >= 100:
            return ContributionTier.SILVER
        return ContributionTier.BRONZE

@dataclass
class User:
    user_id: str
    email: str
    credit_balance: float
    created_at: datetime
    last_active: datetime


@dataclass
class Orchestrator:
    orchestrator_id: str
    host: str
    port: int
    node_count: int
    status: NodeStatus
    registered_at: datetime


@dataclass
class Agent:
    """PicoClaw Agent data model"""
    agent_id: str
    name: str
    node_id: str
    status: AgentStatus
    tools: List[str]
    last_heartbeat: datetime
    created_at: datetime
    config: AgentConfig
    credits_earned: float = 0.0
    credits_spent: float = 0.0
    jobs_completed: int = 0
    error_count: int = 0
    metadata: Dict[str, Any] = None
    
    def __post_init__(self):
        if self.metadata is None:
            self.metadata = {}
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert agent to dictionary for JSON response"""
        return {
            'agent_id': self.agent_id,
            'name': self.name,
            'node_id': self.node_id,
            'status': self.status.value,
            'tools': self.tools,
            'last_heartbeat': self.last_heartbeat.isoformat(),
            'created_at': self.created_at.isoformat(),
            'config': self.config.dict() if isinstance(self.config, AgentConfig) else self.config,
            'credits_earned': self.credits_earned,
            'credits_spent': self.credits_spent,
            'jobs_completed': self.jobs_completed,
            'error_count': self.error_count,
            'metadata': self.metadata
        }


# ============ WebSocket Connection Manager ============

class WebSocketManager:
    """Manages WebSocket connections for real-time updates"""
    
    def __init__(self):
        self.connections: Dict[str, web.WebSocketResponse] = {}  # client_id -> ws
        self.agent_connections: Dict[str, Set[str]] = defaultdict(set)  # agent_id -> set of client_ids
        self.subscriptions: Dict[str, Set[str]] = defaultdict(set)  # client_id -> set of channels
        self._lock = asyncio.Lock()
    
    async def connect(self, client_id: str, ws: web.WebSocketResponse):
        """Register a new WebSocket connection"""
        async with self._lock:
            self.connections[client_id] = ws
            logger.info(f"WebSocket connected: {client_id}")
    
    async def disconnect(self, client_id: str):
        """Remove a WebSocket connection"""
        async with self._lock:
            if client_id in self.connections:
                del self.connections[client_id]
            # Clean up subscriptions
            if client_id in self.subscriptions:
                del self.subscriptions[client_id]
            # Clean up agent connections
            for agent_id in list(self.agent_connections.keys()):
                self.agent_connections[agent_id].discard(client_id)
            logger.info(f"WebSocket disconnected: {client_id}")
    
    async def subscribe(self, client_id: str, channel: str):
        """Subscribe a client to a channel"""
        async with self._lock:
            self.subscriptions[client_id].add(channel)
    
    async def unsubscribe(self, client_id: str, channel: str):
        """Unsubscribe a client from a channel"""
        async with self._lock:
            self.subscriptions[client_id].discard(channel)
    
    async def register_agent_ws(self, agent_id: str, client_id: str):
        """Register WebSocket connection for an agent"""
        async with self._lock:
            self.agent_connections[agent_id].add(client_id)
    
    async def broadcast(self, channel: str, message: Dict[str, Any]):
        """Broadcast message to all subscribers of a channel"""
        message_json = json.dumps(message)
        disconnected = []
        
        for client_id, channels in self.subscriptions.items():
            if channel in channels and client_id in self.connections:
                try:
                    await self.connections[client_id].send_str(message_json)
                except Exception as e:
                    logger.warning(f"Failed to send to {client_id}: {e}")
                    disconnected.append(client_id)
        
        # Clean up disconnected clients
        for client_id in disconnected:
            await self.disconnect(client_id)
    
    async def send_to_agent(self, agent_id: str, message: Dict[str, Any]):
        """Send message to all connections for an agent"""
        message_json = json.dumps(message)
        
        if agent_id not in self.agent_connections:
            return False
        
        for client_id in list(self.agent_connections[agent_id]):
            if client_id in self.connections:
                try:
                    await self.connections[client_id].send_str(message_json)
                except Exception as e:
                    logger.warning(f"Failed to send to agent {agent_id}: {e}")
        
        return True
    
    async def send_to_client(self, client_id: str, message: Dict[str, Any]):
        """Send message to a specific client"""
        if client_id in self.connections:
            try:
                await self.connections[client_id].send_str(json.dumps(message))
                return True
            except Exception as e:
                logger.warning(f"Failed to send to client {client_id}: {e}")
                await self.disconnect(client_id)
        return False
    
    def get_connection_count(self) -> int:
        """Get total number of connections"""
        return len(self.connections)


# ============ Rate Limiter ============

class RateLimiter:
    """Simple in-memory rate limiter"""
    
    def __init__(self):
        self.requests: Dict[str, List[float]] = defaultdict(list)
        self._lock = asyncio.Lock()
    
    async def check_rate_limit(
        self, 
        key: str, 
        max_requests: int = 60, 
        window_seconds: int = 60
    ) -> tuple[bool, int]:
        """
        Check if rate limit is exceeded.
        Returns (is_allowed, remaining_requests)
        """
        async with self._lock:
            now = time.time()
            window_start = now - window_seconds
            
            # Clean old requests
            self.requests[key] = [
                ts for ts in self.requests[key] 
                if ts > window_start
            ]
            
            current_count = len(self.requests[key])
            
            if current_count >= max_requests:
                return False, 0
            
            # Record this request
            self.requests[key].append(now)
            return True, max_requests - current_count - 1


# ============ DEPARROW TOOLS ============

DEPARROW_TOOLS: Dict[str, ToolDefinition] = {
    "job_submit": ToolDefinition(
        name="job_submit",
        description="Submit a compute job to the DEparrow network",
        category="job",
        parameters={
            "spec": {"type": "object", "description": "Job specification"},
            "priority": {"type": "string", "enum": ["low", "normal", "high"], "default": "normal"},
            "timeout": {"type": "integer", "default": 3600}
        },
        rate_limit=30
    ),
    "job_status": ToolDefinition(
        name="job_status",
        description="Get status of a submitted job",
        category="job",
        parameters={
            "job_id": {"type": "string", "description": "Job ID to check"}
        },
        rate_limit=120
    ),
    "job_list": ToolDefinition(
        name="job_list",
        description="List jobs submitted by the agent",
        category="job",
        parameters={
            "status": {"type": "string", "enum": ["pending", "running", "completed", "failed", "all"], "default": "all"},
            "limit": {"type": "integer", "default": 20}
        },
        rate_limit=60
    ),
    "credit_balance": ToolDefinition(
        name="credit_balance",
        description="Get current credit balance",
        category="credit",
        parameters={},
        rate_limit=120
    ),
    "credit_transfer": ToolDefinition(
        name="credit_transfer",
        description="Transfer credits to another user or agent",
        category="credit",
        parameters={
            "to": {"type": "string", "description": "Recipient user/agent ID"},
            "amount": {"type": "number", "description": "Amount to transfer"}
        },
        rate_limit=30
    ),
    "credit_history": ToolDefinition(
        name="credit_history",
        description="Get credit transaction history",
        category="credit",
        parameters={
            "limit": {"type": "integer", "default": 50}
        },
        rate_limit=60
    ),
    "node_status": ToolDefinition(
        name="node_status",
        description="Get status of a compute node",
        category="node",
        parameters={
            "node_id": {"type": "string", "description": "Node ID to check"}
        },
        rate_limit=120
    ),
    "node_list": ToolDefinition(
        name="node_list",
        description="List available compute nodes",
        category="node",
        parameters={
            "status": {"type": "string", "enum": ["online", "offline", "all"], "default": "online"},
            "limit": {"type": "integer", "default": 50}
        },
        rate_limit=60
    ),
    "node_contribution": ToolDefinition(
        name="node_contribution",
        description="Get contribution stats for a node",
        category="node",
        parameters={
            "node_id": {"type": "string", "description": "Node ID to check"}
        },
        rate_limit=60
    ),
    "wallet_create": ToolDefinition(
        name="wallet_create",
        description="Create a new wallet address",
        category="wallet",
        parameters={},
        rate_limit=10
    ),
    "wallet_balance": ToolDefinition(
        name="wallet_balance",
        description="Get wallet balance",
        category="wallet",
        parameters={},
        rate_limit=120
    ),
    "wallet_transactions": ToolDefinition(
        name="wallet_transactions",
        description="Get wallet transaction history",
        category="wallet",
        parameters={
            "limit": {"type": "integer", "default": 50}
        },
        rate_limit=60
    ),
    "agent_self_terminate": ToolDefinition(
        name="agent_self_terminate",
        description="Terminate current agent instance (requires confirmation)",
        category="agent",
        parameters={
            "confirm": {"type": "boolean", "description": "Confirmation flag"}
        },
        rate_limit=5
    ),
    "agent_spawn": ToolDefinition(
        name="agent_spawn",
        description="Spawn a new agent instance",
        category="agent",
        parameters={
            "name": {"type": "string", "description": "Name for new agent"},
            "config": {"type": "object", "description": "Agent configuration"}
        },
        rate_limit=10
    )
}

class DEparrowBootstrapServer:
    """DEparrow Bootstrap Server - Meta-OS Control Plane"""
    
    def __init__(self):
        self.app = web.Application()
        self.setup_routes()
        self.db_pool = None
        self.redis_client = None
        self.session = None
        
        # WebSocket manager
        self.ws_manager = WebSocketManager()
        
        # Rate limiter
        self.rate_limiter = RateLimiter()
        
        # In-memory caches
        self.nodes: Dict[str, Node] = {}
        self.orchestrators: Dict[str, Orchestrator] = {}
        self.users: Dict[str, User] = {}
        self.jobs: Dict[str, JobSubmission] = {}
        
        # PicoClaw Agents
        self.agents: Dict[str, Agent] = {}
        self.agent_tool_executions: Dict[str, List[ToolExecutionResult]] = defaultdict(list)
        
        # Initialize with sample data
        self._initialize_sample_data()
    
    def _initialize_sample_data(self):
        """Initialize with sample data for testing"""
        # Sample orchestrator
        self.orchestrators["orch-1"] = Orchestrator(
            orchestrator_id="orch-1",
            host="orchestrator.deparrow.net",
            port=4222,
            node_count=0,
            status=NodeStatus.ONLINE,
            registered_at=datetime.utcnow()
        )
        
        # Sample user
        self.users["user-1"] = User(
            user_id="user-1",
            email="user@example.com",
            credit_balance=100.0,
            created_at=datetime.utcnow(),
            last_active=datetime.utcnow()
        )
        
        # Sample agent
        self.agents["agent-demo-1"] = Agent(
            agent_id="agent-demo-1",
            name="Demo Agent",
            node_id="node-demo-1",
            status=AgentStatus.IDLE,
            tools=["job", "credit", "node", "wallet"],
            last_heartbeat=datetime.utcnow(),
            created_at=datetime.utcnow(),
            config=AgentConfig()
        )
    
    def setup_routes(self):
        """Setup API routes"""
        # Node management routes
        self.app.router.add_post('/api/v1/nodes/register', self.register_node)
        self.app.router.add_get('/api/v1/nodes', self.list_nodes)
        self.app.router.add_get('/api/v1/nodes/{node_id}', self.get_node)
        self.app.router.add_post('/api/v1/nodes/{node_id}/heartbeat', self.node_heartbeat)
        
        # Contribution tracking routes
        self.app.router.add_get('/api/v1/nodes/{node_id}/contribution', self.get_node_contribution)
        self.app.router.add_get('/api/v1/network/contribution', self.get_network_contribution)
        self.app.router.add_get('/api/v1/network/leaderboard', self.get_leaderboard)
        self.app.router.add_get('/api/v1/network/globe', self.get_globe_data)
        
        # Orchestrator routes
        self.app.router.add_post('/api/v1/orchestrators/register', self.register_orchestrator)
        self.app.router.add_get('/api/v1/orchestrators', self.list_orchestrators)
        
        # Job routes
        self.app.router.add_post('/api/v1/jobs/submit', self.submit_job)
        self.app.router.add_get('/api/v1/jobs/{job_id}', self.get_job_status)
        self.app.router.add_post('/api/v1/jobs/{job_id}/cancel', self.cancel_job)
        
        # Credit routes
        self.app.router.add_post('/api/v1/credits/check', self.check_credits)
        self.app.router.add_post('/api/v1/credits/transfer', self.transfer_credits)
        self.app.router.add_get('/api/v1/credits/balance/{user_id}', self.get_credit_balance)
        
        # ============ PicoClaw Agent Routes ============
        # Agent management
        self.app.router.add_post('/api/v1/agent/register', self.register_agent)
        self.app.router.add_get('/api/v1/agent/{agent_id}', self.get_agent)
        self.app.router.add_get('/api/v1/agents', self.list_agents)
        self.app.router.add_put('/api/v1/agent/{agent_id}/config', self.update_agent_config)
        self.app.router.add_post('/api/v1/agent/{agent_id}/heartbeat', self.agent_heartbeat)
        self.app.router.add_delete('/api/v1/agent/{agent_id}', self.delete_agent)
        
        # Tool endpoints
        self.app.router.add_get('/api/v1/tools', self.list_tools)
        self.app.router.add_get('/api/v1/tools/{tool_name}', self.get_tool_info)
        self.app.router.add_post('/api/v1/tools/{tool_name}/execute', self.execute_tool)
        
        # WebSocket endpoint
        self.app.router.add_get('/api/v1/ws', self.websocket_handler)
        
        # Health and metrics
        self.app.router.add_get('/api/v1/health', self.health_check)
        self.app.router.add_get('/api/v1/metrics', self.get_metrics)
    
    # Authentication middleware
    @web.middleware
    async def auth_middleware(self, request: web.Request, handler):
        """JWT authentication middleware"""
        # Skip auth for public endpoints
        public_paths = ['/api/v1/health', '/api/v1/nodes/register', '/api/v1/orchestrators/register']
        if request.path in public_paths:
            return await handler(request)
        
        # Check for Authorization header
        auth_header = request.headers.get('Authorization')
        if not auth_header or not auth_header.startswith('Bearer '):
            return web.json_response(
                {'error': 'Missing or invalid authorization token'},
                status=401
            )
        
        token = auth_header[7:]  # Remove 'Bearer ' prefix
        
        try:
            # Verify JWT token
            payload = jwt.decode(
                token,
                Config.JWT_SECRET,
                algorithms=[Config.JWT_ALGORITHM]
            )
            request['user_id'] = payload.get('user_id')
            request['role'] = payload.get('role', 'user')
        except jwt.ExpiredSignatureError:
            return web.json_response({'error': 'Token expired'}, status=401)
        except jwt.InvalidTokenError:
            return web.json_response({'error': 'Invalid token'}, status=401)
        
        return await handler(request)
    
    # Node Management
    async def register_node(self, request: web.Request):
        """Register a new compute node with DEparrow network"""
        try:
            data = await request.json()
            registration = NodeRegistration(**data)
            
            # Check if node already exists
            if registration.node_id in self.nodes:
                node = self.nodes[registration.node_id]
                node.last_seen = datetime.utcnow()
                node.status = NodeStatus.ONLINE
                logger.info(f"Node re-registered: {registration.node_id}")
            else:
                # Create new node
                node = Node(
                    node_id=registration.node_id,
                    public_key=registration.public_key,
                    arch=registration.arch,
                    resources=registration.resources,
                    status=NodeStatus.ONLINE,
                    last_seen=datetime.utcnow(),
                    labels=registration.labels
                )
                self.nodes[registration.node_id] = node
                logger.info(f"New node registered: {registration.node_id}")
            
            # Return orchestrator information
            orchestrators = [
                {
                    'host': orch.host,
                    'port': orch.port,
                    'orchestrator_id': orch.orchestrator_id
                }
                for orch in self.orchestrators.values()
                if orch.status == NodeStatus.ONLINE
            ]
            
            # Generate JWT token for node
            node_token = jwt.encode(
                {
                    'node_id': registration.node_id,
                    'role': 'node',
                    'exp': datetime.utcnow() + timedelta(hours=Config.JWT_EXPIRY_HOURS)
                },
                Config.JWT_SECRET,
                algorithm=Config.JWT_ALGORITHM
            )
            
            response = {
                'status': 'registered',
                'node_id': registration.node_id,
                'orchestrators': orchestrators,
                'token': node_token,
                'credit_earning_rate': Config.CREDIT_EARNING_RATE,
                'message': 'Node registered successfully with DEparrow network'
            }
            
            return web.json_response(response, status=200)
            
        except Exception as e:
            logger.error(f"Node registration error: {str(e)}")
            return web.json_response(
                {'error': f'Registration failed: {str(e)}'},
                status=400
            )
    
    async def list_nodes(self, request: web.Request):
        """List all registered nodes"""
        nodes_list = []
        for node_id, node in self.nodes.items():
            nodes_list.append({
                'node_id': node_id,
                'arch': node.arch,
                'status': node.status,
                'last_seen': node.last_seen.isoformat(),
                'resources': node.resources,
                'credits_earned': node.credits_earned,
                'labels': node.labels
            })
        
        return web.json_response({
            'nodes': nodes_list,
            'total': len(nodes_list),
            'online': len([n for n in self.nodes.values() if n.status == NodeStatus.ONLINE])
        })
    
    async def get_node(self, request: web.Request):
        """Get node details"""
        node_id = request.match_info['node_id']
        
        if node_id not in self.nodes:
            return web.json_response({'error': 'Node not found'}, status=404)
        
        node = self.nodes[node_id]
        return web.json_response(asdict(node))
    
    async def node_heartbeat(self, request: web.Request):
        """Node heartbeat to maintain registration"""
        node_id = request.match_info['node_id']
        
        if node_id not in self.nodes:
            return web.json_response({'error': 'Node not found'}, status=404)
        
        node = self.nodes[node_id]
        node.last_seen = datetime.utcnow()
        
        # Update credits based on resource usage
        # In production, this would calculate based on actual usage
        node.credits_earned += Config.CREDIT_EARNING_RATE * 0.1  # Sample calculation
        
        return web.json_response({
            'status': 'ok',
            'last_seen': node.last_seen.isoformat(),
            'credits_earned': node.credits_earned
        })
    
    # Orchestrator Management
    async def register_orchestrator(self, request: web.Request):
        """Register a new orchestrator"""
        try:
            data = await request.json()
            orchestrator_id = data.get('orchestrator_id')
            host = data.get('host')
            port = data.get('port', Config.ORCHESTRATOR_PORT)
            
            if not orchestrator_id or not host:
                return web.json_response(
                    {'error': 'Missing required fields'},
                    status=400
                )
            
            # Create or update orchestrator
            orchestrator = Orchestrator(
                orchestrator_id=orchestrator_id,
                host=host,
                port=port,
                node_count=0,
                status=NodeStatus.ONLINE,
                registered_at=datetime.utcnow()
            )
            
            self.orchestrators[orchestrator_id] = orchestrator
            logger.info(f"Orchestrator registered: {orchestrator_id} at {host}:{port}")
            
            # Generate JWT token for orchestrator
            orch_token = jwt.encode(
                {
                    'orchestrator_id': orchestrator_id,
                    'role': 'orchestrator',
                    'exp': datetime.utcnow() + timedelta(hours=Config.JWT_EXPIRY_HOURS)
                },
                Config.JWT_SECRET,
                algorithm=Config.JWT_ALGORITHM
            )
            
            return web.json_response({
                'status': 'registered',
                'orchestrator_id': orchestrator_id,
                'token': orch_token,
                'message': 'Orchestrator registered with DEparrow bootstrap'
            })
            
        except Exception as e:
            logger.error(f"Orchestrator registration error: {str(e)}")
            return web.json_response(
                {'error': f'Registration failed: {str(e)}'},
                status=400
            )
    
    async def list_orchestrators(self, request: web.Request):
        """List all registered orchestrators"""
        orchestrators_list = []
        for orch_id, orchestrator in self.orchestrators.items():
            orchestrators_list.append({
                'orchestrator_id': orch_id,
                'host': orchestrator.host,
                'port': orchestrator.port,
                'status': orchestrator.status,
                'node_count': orchestrator.node_count,
                'registered_at': orchestrator.registered_at.isoformat()
            })
        
        return web.json_response({
            'orchestrators': orchestrators_list,
            'total': len(orchestrators_list)
        })
    
    # Job Management with Credit System
    async def submit_job(self, request: web.Request):
        """Submit a job with credit payment verification"""
        try:
            data = await request.json()
            user_id = request['user_id']
            
            # Check user credit balance
            if user_id not in self.users:
                return web.json_response({'error': 'User not found'}, status=404)
            
            user = self.users[user_id]
            credit_cost = data.get('credit_cost', Config.CREDIT_SUBMISSION_COST)
            
            if user.credit_balance < credit_cost:
                return web.json_response({
                    'error': 'Insufficient credits',
                    'required': credit_cost,
                    'available': user.credit_balance
                }, status=402)  # Payment Required
            
            # Deduct credits
            user.credit_balance -= credit_cost
            user.last_active = datetime.utcnow()
            
            # Create job record
            job_id = data.get('job_id', f"job-{datetime.utcnow().timestamp()}")
            job = JobSubmission(
                job_id=job_id,
                user_id=user_id,
                spec=data.get('spec', {}),
                credit_cost=credit_cost,
                orchestrator=data.get('orchestrator')
            )
            
            self.jobs[job_id] = job
            logger.info(f"Job submitted: {job_id} by user {user_id}, cost: {credit_cost} credits")
            
            # In production, this would forward to the actual orchestrator
            # For now, simulate job acceptance
            
            return web.json_response({
                'status': 'accepted',
                'job_id': job_id,
                'credit_deducted': credit_cost,
                'remaining_balance': user.credit_balance,
                'message': 'Job submitted successfully. Credits deducted.'
            })
            
        except Exception as e:
            logger.error(f"Job submission error: {str(e)}")
            return web.json_response(
                {'error': f'Job submission failed: {str(e)}'},
                status=400
            )
    
    async def get_job_status(self, request: web.Request):
        """Get job status"""
        job_id = request.match_info['job_id']
        
        if job_id not in self.jobs:
            return web.json_response({'error': 'Job not found'}, status=404)
        
        job = self.jobs[job_id]
        
        # Simulate job status
        status = JobStatus.RUNNING  # In production, get from orchestrator
        
        return web.json_response({
            'job_id': job_id,
            'status': status,
            'user_id': job.user_id,
            'credit_cost': job.credit_cost,
            'submitted_at': datetime.utcnow().isoformat()  # Would be actual timestamp
        })
    
    async def cancel_job(self, request: web.Request):
        """Cancel a job and refund credits"""
        job_id = request.match_info['job_id']
        user_id = request['user_id']
        
        if job_id not in self.jobs:
            return web.json_response({'error': 'Job not found'}, status=404)
        
        job = self.jobs[job_id]
        
        # Check ownership
        if job.user_id != user_id:
            return web.json_response({'error': 'Not authorized'}, status=403)
        
        # Refund credits (partial or full based on job progress)
        refund_amount = job.credit_cost * 0.5  # 50% refund for cancellation
        
        if user_id in self.users:
            user = self.users[user_id]
            user.credit_balance += refund_amount
        
        # Update job status
        # In production, would notify orchestrator
        
        logger.info(f"Job cancelled: {job_id}, refunded: {refund_amount} credits")
        
        return web.json_response({
            'status': 'cancelled',
            'job_id': job_id,
            'refund_amount': refund_amount,
            'remaining_balance': user.credit_balance if user_id in self.users else 0
        })
    
    # Credit Management
    async def check_credits(self, request: web.Request):
        """Check if user has sufficient credits for an operation"""
        user_id = request['user_id']
        data = await request.json()
        required = data.get('required', Config.CREDIT_SUBMISSION_COST)
        
        if user_id not in self.users:
            return web.json_response({'error': 'User not found'}, status=404)
        
        user = self.users[user_id]
        has_sufficient = user.credit_balance >= required
        
        return web.json_response({
            'has_sufficient': has_sufficient,
            'required': required,
            'available': user.credit_balance,
            'difference': user.credit_balance - required
        })
    
    async def transfer_credits(self, request: web.Request):
        """Transfer credits between users"""
        try:
            data = await request.json()
            from_user_id = request['user_id']
            to_user_id = data.get('to_user_id')
            amount = data.get('amount')
            
            if not to_user_id or not amount or amount <= 0:
                return web.json_response(
                    {'error': 'Invalid transfer request'},
                    status=400
                )
            
            # Check if both users exist
            if from_user_id not in self.users or to_user_id not in self.users:
                return web.json_response({'error': 'User not found'}, status=404)
            
            from_user = self.users[from_user_id]
            to_user = self.users[to_user_id]
            
            # Check balance
            if from_user.credit_balance < amount:
                return web.json_response({
                    'error': 'Insufficient credits for transfer',
                    'available': from_user.credit_balance,
                    'required': amount
                }, status=402)
            
            # Perform transfer
            from_user.credit_balance -= amount
            to_user.credit_balance += amount
            
            logger.info(f"Credit transfer: {amount} from {from_user_id} to {to_user_id}")
            
            return web.json_response({
                'status': 'transferred',
                'amount': amount,
                'from_user': from_user_id,
                'to_user': to_user_id,
                'from_balance': from_user.credit_balance,
                'to_balance': to_user.credit_balance
            })
            
        except Exception as e:
            logger.error(f"Credit transfer error: {str(e)}")
            return web.json_response(
                {'error': f'Transfer failed: {str(e)}'},
                status=400
            )
    
    async def get_credit_balance(self, request: web.Request):
        """Get user credit balance"""
        user_id = request.match_info['user_id']
        
        if user_id not in self.users:
            return web.json_response({'error': 'User not found'}, status=404)
        
        user = self.users[user_id]
        
        return web.json_response({
            'user_id': user_id,
            'credit_balance': user.credit_balance,
            'last_active': user.last_active.isoformat()
        })
    
    # Health and Metrics
    async def health_check(self, request: web.Request):
        """Health check endpoint"""
        return web.json_response({
            'status': 'healthy',
            'timestamp': datetime.utcnow().isoformat(),
            'version': '1.1.0',
            'components': {
                'nodes': len(self.nodes),
                'orchestrators': len(self.orchestrators),
                'users': len(self.users),
                'jobs': len(self.jobs),
                'agents': len(self.agents),
                'ws_connections': self.ws_manager.get_connection_count()
            }
        })
    
    async def get_metrics(self, request: web.Request):
        """Get system metrics"""
        online_nodes = len([n for n in self.nodes.values() if n.status == NodeStatus.ONLINE])
        total_credits = sum(user.credit_balance for user in self.users.values())
        node_credits = sum(node.credits_earned for node in self.nodes.values())
        
        # Agent metrics
        online_agents = len([a for a in self.agents.values() if a.status != AgentStatus.OFFLINE])
        working_agents = len([a for a in self.agents.values() if a.status == AgentStatus.WORKING])
        agent_jobs_completed = sum(a.jobs_completed for a in self.agents.values())
        agent_credits_earned = sum(a.credits_earned for a in self.agents.values())
        
        return web.json_response({
            'metrics': {
                'nodes': {
                    'total': len(self.nodes),
                    'online': online_nodes,
                    'by_arch': {
                        arch: len([n for n in self.nodes.values() if n.arch == arch])
                        for arch in NodeArchitecture
                    }
                },
                'credits': {
                    'total_circulating': total_credits,
                    'total_earned': node_credits,
                    'user_balances': {
                        user_id: user.credit_balance
                        for user_id, user in self.users.items()
                    }
                },
                'orchestrators': {
                    'total': len(self.orchestrators),
                    'online': len([o for o in self.orchestrators.values() if o.status == NodeStatus.ONLINE])
                },
                'jobs': {
                    'total': len(self.jobs),
                    'active': len([j for j in self.jobs.values()])
                },
                'agents': {
                    'total': len(self.agents),
                    'online': online_agents,
                    'working': working_agents,
                    'jobs_completed': agent_jobs_completed,
                    'credits_earned': agent_credits_earned,
                    'by_status': {
                        status: len([a for a in self.agents.values() if a.status == status])
                        for status in AgentStatus
                    }
                },
                'websocket': {
                    'connections': self.ws_manager.get_connection_count()
                }
            },
            'timestamp': datetime.utcnow().isoformat()
        })
    
    # ============ Contribution Tracking ============
    
    async def get_node_contribution(self, request: web.Request):
        """Get a node's contribution percentage to the network"""
        node_id = request.match_info['node_id']
        
        if node_id not in self.nodes:
            return web.json_response({'error': 'Node not found'}, status=404)
        
        node = self.nodes[node_id]
        
        # Calculate network totals
        total_cpu_hours = sum(n.cpu_usage_hours for n in self.nodes.values())
        total_gpu_hours = sum(n.gpu_usage_hours for n in self.nodes.values())
        total_gflops = sum(n.live_gflops for n in self.nodes.values())
        
        # Calculate percentages
        cpu_percent = (node.cpu_usage_hours / total_cpu_hours * 100) if total_cpu_hours > 0 else 0
        gpu_percent = (node.gpu_usage_hours / total_gpu_hours * 100) if total_gpu_hours > 0 else 0
        gflops_percent = (node.live_gflops / total_gflops * 100) if total_gflops > 0 else 0
        
        # Calculate rank
        sorted_nodes = sorted(
            self.nodes.values(),
            key=lambda n: n.cpu_usage_hours + n.gpu_usage_hours,
            reverse=True
        )
        rank = next((i + 1 for i, n in enumerate(sorted_nodes) if n.node_id == node_id), 0)
        
        return web.json_response({
            'node_id': node_id,
            'contribution': {
                'cpu': {
                    'cores': node.cpu_cores,
                    'usage_hours': node.cpu_usage_hours,
                    'percent_of_network': round(cpu_percent, 2)
                },
                'gpu': {
                    'count': node.gpu_count,
                    'model': node.gpu_model,
                    'usage_hours': node.gpu_usage_hours,
                    'percent_of_network': round(gpu_percent, 2)
                },
                'memory_gb': node.memory_gb,
                'live_gflops': node.live_gflops,
                'gflops_percent': round(gflops_percent, 2)
            },
            'ranking': {
                'rank': rank,
                'total_nodes': len(self.nodes),
                'tier': node.get_tier().value,
                'tier_icon': self._get_tier_icon(node.get_tier())
            },
            'credits_earned': node.credits_earned,
            'pulse': node.status == NodeStatus.ONLINE,  # For animation
            'timestamp': datetime.utcnow().isoformat()
        })
    
    def _get_tier_icon(self, tier: ContributionTier) -> str:
        """Get emoji icon for tier"""
        icons = {
            ContributionTier.BRONZE: "🥉",
            ContributionTier.SILVER: "🥈",
            ContributionTier.GOLD: "🥇",
            ContributionTier.DIAMOND: "💎",
            ContributionTier.LEGENDARY: "🔥"
        }
        return icons.get(tier, "🥉")
    
    async def get_network_contribution(self, request: web.Request):
        """Get total network contribution statistics"""
        online_nodes = [n for n in self.nodes.values() if n.status == NodeStatus.ONLINE]
        
        total_cpu_cores = sum(n.cpu_cores for n in self.nodes.values())
        total_cpu_hours = sum(n.cpu_usage_hours for n in self.nodes.values())
        total_gpu_count = sum(n.gpu_count for n in self.nodes.values())
        total_gpu_hours = sum(n.gpu_usage_hours for n in self.nodes.values())
        total_memory_gb = sum(n.memory_gb for n in self.nodes.values())
        live_gflops = sum(n.live_gflops for n in online_nodes)
        
        # Tier distribution
        tier_counts = {}
        for tier in ContributionTier:
            tier_counts[tier.value] = len([n for n in self.nodes.values() if n.get_tier() == tier])
        
        return web.json_response({
            'network': {
                'total_nodes': len(self.nodes),
                'online_nodes': len(online_nodes),
                'total_cpu_cores': total_cpu_cores,
                'total_cpu_hours': round(total_cpu_hours, 2),
                'total_gpu_count': total_gpu_count,
                'total_gpu_hours': round(total_gpu_hours, 2),
                'total_memory_gb': round(total_memory_gb, 2),
                'live_gflops': round(live_gflops, 2),
                'live_tflops': round(live_gflops / 1000, 2)
            },
            'tiers': tier_counts,
            'pulse': True,  # Network is alive
            'timestamp': datetime.utcnow().isoformat()
        })
    
    async def get_leaderboard(self, request: web.Request):
        """Get contribution leaderboard with rankings"""
        limit = int(request.query.get('limit', 20))
        
        # Sort by total contribution
        sorted_nodes = sorted(
            self.nodes.values(),
            key=lambda n: n.cpu_usage_hours + n.gpu_usage_hours,
            reverse=True
        )[:limit]
        
        leaderboard = []
        for rank, node in enumerate(sorted_nodes, 1):
            leaderboard.append({
                'rank': rank,
                'node_id': node.node_id[:12] + '...',  # Truncate for privacy
                'tier': node.get_tier().value,
                'tier_icon': self._get_tier_icon(node.get_tier()),
                'cpu_hours': round(node.cpu_usage_hours, 2),
                'gpu_hours': round(node.gpu_usage_hours, 2),
                'total_hours': round(node.cpu_usage_hours + node.gpu_usage_hours, 2),
                'live_gflops': round(node.live_gflops, 2),
                'credits_earned': round(node.credits_earned, 2),
                'status': node.status.value,
                'pulse': node.status == NodeStatus.ONLINE
            })
        
        return web.json_response({
            'leaderboard': leaderboard,
            'total_participants': len(self.nodes),
            'timestamp': datetime.utcnow().isoformat()
        })
    
    async def get_globe_data(self, request: web.Request):
        """Get node locations for 3D globe visualization"""
        globe_nodes = []
        
        for node in self.nodes.values():
            if node.location.get('lat') and node.location.get('lng'):
                globe_nodes.append({
                    'id': node.node_id[:8],
                    'lat': node.location['lat'],
                    'lng': node.location['lng'],
                    'status': node.status.value,
                    'gflops': node.live_gflops,
                    'tier': node.get_tier().value,
                    'pulse': node.status == NodeStatus.ONLINE
                })
        
        # Calculate connection lines (top contributors connect to each other)
        connections = []
        top_nodes = sorted(globe_nodes, key=lambda n: n['gflops'], reverse=True)[:10]
        for i, node in enumerate(top_nodes[:-1]):
            connections.append({
                'from': {'lat': node['lat'], 'lng': node['lng']},
                'to': {'lat': top_nodes[i + 1]['lat'], 'lng': top_nodes[i + 1]['lng']},
                'strength': min(node['gflops'], top_nodes[i + 1]['gflops'])
            })
        
        return web.json_response({
            'nodes': globe_nodes,
            'connections': connections,
            'center': {'lat': 20.0, 'lng': 0.0},
            'timestamp': datetime.utcnow().isoformat()
        })
    
    # ============ PicoClaw Agent Endpoints ============
    
    async def register_agent(self, request: web.Request):
        """Register a new PicoClaw agent"""
        try:
            data = await request.json()
            registration = AgentRegistration(**data)
            
            # Verify node exists (optional - can create virtual node)
            if registration.node_id not in self.nodes and registration.node_id != "virtual":
                # Create a virtual node for the agent
                self.nodes[registration.node_id] = Node(
                    node_id=registration.node_id,
                    public_key=f"agent-{registration.agent_id}",
                    arch=NodeArchitecture.X86_64,
                    resources={"cpu": 1, "memory": "1GB"},
                    status=NodeStatus.ONLINE,
                    last_seen=datetime.utcnow(),
                    labels={"type": "agent", "agent_id": registration.agent_id}
                )
            
            # Create agent
            agent = Agent(
                agent_id=registration.agent_id,
                name=registration.name,
                node_id=registration.node_id,
                status=AgentStatus.INITIALIZING,
                tools=registration.config.tools_enabled,
                last_heartbeat=datetime.utcnow(),
                created_at=datetime.utcnow(),
                config=registration.config
            )
            
            self.agents[registration.agent_id] = agent
            logger.info(f"Agent registered: {registration.agent_id} ({registration.name})")
            
            # Generate JWT token for agent
            agent_token = jwt.encode(
                {
                    'agent_id': registration.agent_id,
                    'node_id': registration.node_id,
                    'role': 'agent',
                    'tools': registration.config.tools_enabled,
                    'exp': datetime.utcnow() + timedelta(hours=Config.JWT_EXPIRY_HOURS * 7)  # 7 days for agents
                },
                Config.JWT_SECRET,
                algorithm=Config.JWT_ALGORITHM
            )
            
            # Broadcast agent registration
            await self.ws_manager.broadcast('agents', {
                'type': 'agent_registered',
                'agent_id': registration.agent_id,
                'name': registration.name,
                'timestamp': datetime.utcnow().isoformat()
            })
            
            return web.json_response({
                'status': 'registered',
                'agent_id': registration.agent_id,
                'token': agent_token,
                'tools': registration.config.tools_enabled,
                'message': 'Agent registered successfully'
            })
            
        except ValidationError as e:
            return web.json_response({'error': str(e)}, status=400)
        except Exception as e:
            logger.error(f"Agent registration error: {str(e)}")
            return web.json_response({'error': f'Registration failed: {str(e)}'}, status=400)
    
    async def get_agent(self, request: web.Request):
        """Get agent status and details"""
        agent_id = request.match_info['agent_id']
        
        if agent_id not in self.agents:
            return web.json_response({'error': 'Agent not found'}, status=404)
        
        agent = self.agents[agent_id]
        
        # Get recent tool executions
        recent_executions = self.agent_tool_executions.get(agent_id, [])[-10:]
        
        return web.json_response({
            'agent': agent.to_dict(),
            'recent_executions': [e.dict() for e in recent_executions],
            'timestamp': datetime.utcnow().isoformat()
        })
    
    async def list_agents(self, request: web.Request):
        """List all registered agents"""
        status_filter = request.query.get('status')
        
        agents_list = []
        for agent_id, agent in self.agents.items():
            if status_filter and agent.status.value != status_filter:
                continue
            agents_list.append({
                'agent_id': agent_id,
                'name': agent.name,
                'status': agent.status.value,
                'node_id': agent.node_id,
                'tools': agent.tools,
                'last_heartbeat': agent.last_heartbeat.isoformat(),
                'credits_earned': agent.credits_earned,
                'jobs_completed': agent.jobs_completed
            })
        
        return web.json_response({
            'agents': agents_list,
            'total': len(agents_list),
            'online': len([a for a in self.agents.values() if a.status in [AgentStatus.IDLE, AgentStatus.WORKING]])
        })
    
    async def update_agent_config(self, request: web.Request):
        """Update agent configuration"""
        agent_id = request.match_info['agent_id']
        
        if agent_id not in self.agents:
            return web.json_response({'error': 'Agent not found'}, status=404)
        
        try:
            data = await request.json()
            agent = self.agents[agent_id]
            
            # Update config fields
            if 'model' in data:
                agent.config.model = data['model']
            if 'workspace' in data:
                agent.config.workspace = data['workspace']
            if 'max_iterations' in data:
                agent.config.max_iterations = data['max_iterations']
            if 'auto_approve' in data:
                agent.config.auto_approve = data['auto_approve']
            if 'tools_enabled' in data:
                agent.config.tools_enabled = data['tools_enabled']
                agent.tools = data['tools_enabled']
            if 'temperature' in data:
                agent.config.temperature = data['temperature']
            if 'system_prompt' in data:
                agent.config.system_prompt = data['system_prompt']
            
            logger.info(f"Agent config updated: {agent_id}")
            
            # Notify agent via WebSocket
            await self.ws_manager.send_to_agent(agent_id, {
                'type': 'config_update',
                'config': agent.config.dict(),
                'timestamp': datetime.utcnow().isoformat()
            })
            
            return web.json_response({
                'status': 'updated',
                'agent_id': agent_id,
                'config': agent.config.dict()
            })
            
        except Exception as e:
            logger.error(f"Config update error: {str(e)}")
            return web.json_response({'error': str(e)}, status=400)
    
    async def agent_heartbeat(self, request: web.Request):
        """Agent heartbeat to maintain registration"""
        agent_id = request.match_info['agent_id']
        
        if agent_id not in self.agents:
            return web.json_response({'error': 'Agent not found'}, status=404)
        
        try:
            data = await request.json() if request.content_length else {}
            agent = self.agents[agent_id]
            agent.last_heartbeat = datetime.utcnow()
            
            # Update status if provided
            if 'status' in data:
                agent.status = AgentStatus(data['status'])
            if 'jobs_completed' in data:
                agent.jobs_completed = data['jobs_completed']
            if 'credits_earned' in data:
                agent.credits_earned = data['credits_earned']
            
            return web.json_response({
                'status': 'ok',
                'timestamp': datetime.utcnow().isoformat(),
                'config': agent.config.dict() if agent.config else None
            })
            
        except Exception as e:
            logger.error(f"Heartbeat error: {str(e)}")
            return web.json_response({'error': str(e)}, status=400)
    
    async def delete_agent(self, request: web.Request):
        """Delete/deregister an agent"""
        agent_id = request.match_info['agent_id']
        
        if agent_id not in self.agents:
            return web.json_response({'error': 'Agent not found'}, status=404)
        
        agent = self.agents[agent_id]
        
        # Broadcast termination
        await self.ws_manager.broadcast('agents', {
            'type': 'agent_terminated',
            'agent_id': agent_id,
            'name': agent.name,
            'timestamp': datetime.utcnow().isoformat()
        })
        
        # Remove agent
        del self.agents[agent_id]
        if agent_id in self.agent_tool_executions:
            del self.agent_tool_executions[agent_id]
        
        logger.info(f"Agent deleted: {agent_id}")
        
        return web.json_response({
            'status': 'deleted',
            'agent_id': agent_id
        })
    
    # ============ Tool Endpoints ============
    
    async def list_tools(self, request: web.Request):
        """List all available DEparrow tools"""
        category = request.query.get('category')
        
        tools_list = []
        for name, tool in DEPARROW_TOOLS.items():
            if category and tool.category != category:
                continue
            tools_list.append({
                'name': name,
                'description': tool.description,
                'category': tool.category,
                'parameters': tool.parameters,
                'requires_auth': tool.requires_auth,
                'rate_limit': tool.rate_limit
            })
        
        return web.json_response({
            'tools': tools_list,
            'total': len(tools_list),
            'categories': list(set(t.category for t in DEPARROW_TOOLS.values()))
        })
    
    async def get_tool_info(self, request: web.Request):
        """Get detailed information about a specific tool"""
        tool_name = request.match_info['tool_name']
        
        if tool_name not in DEPARROW_TOOLS:
            return web.json_response({'error': 'Tool not found'}, status=404)
        
        tool = DEPARROW_TOOLS[tool_name]
        
        return web.json_response({
            'name': tool.name,
            'description': tool.description,
            'category': tool.category,
            'parameters': tool.parameters,
            'requires_auth': tool.requires_auth,
            'rate_limit': tool.rate_limit,
            'example_usage': self._get_tool_example(tool_name)
        })
    
    def _get_tool_example(self, tool_name: str) -> Dict[str, Any]:
        """Get example usage for a tool"""
        examples = {
            "job_submit": {
                "spec": {"image": "python:3.11", "command": "python -c 'print(\"Hello\")'"},
                "priority": "normal"
            },
            "credit_balance": {},
            "node_list": {"status": "online"},
            "wallet_balance": {}
        }
        return examples.get(tool_name, {})
    
    async def execute_tool(self, request: web.Request):
        """Execute a tool via API"""
        tool_name = request.match_info['tool_name']
        
        if tool_name not in DEPARROW_TOOLS:
            return web.json_response({'error': 'Tool not found'}, status=404)
        
        tool = DEPARROW_TOOLS[tool_name]
        
        try:
            data = await request.json()
            agent_id = data.get('agent_id', request.get('user_id', 'unknown'))
            parameters = data.get('parameters', {})
            
            # Rate limiting
            rate_key = f"{agent_id}:{tool_name}"
            allowed, remaining = await self.rate_limiter.check_rate_limit(
                rate_key, 
                max_requests=tool.rate_limit
            )
            
            if not allowed:
                return web.json_response({
                    'error': 'Rate limit exceeded',
                    'tool': tool_name,
                    'limit': tool.rate_limit
                }, status=429)
            
            # Execute tool
            start_time = time.time()
            result = await self._execute_tool_impl(tool_name, parameters, agent_id)
            duration_ms = (time.time() - start_time) * 1000
            
            execution_result = ToolExecutionResult(
                execution_id=f"exec-{uuid.uuid4().hex[:8]}",
                tool_name=tool_name,
                status=result.get('status', 'success'),
                result=result.get('data'),
                error=result.get('error'),
                duration_ms=duration_ms
            )
            
            # Store execution history
            self.agent_tool_executions[agent_id].append(execution_result)
            
            # Broadcast tool execution
            await self.ws_manager.broadcast('tool_executions', {
                'type': 'tool_executed',
                'agent_id': agent_id,
                'tool_name': tool_name,
                'status': execution_result.status,
                'duration_ms': duration_ms,
                'timestamp': datetime.utcnow().isoformat()
            })
            
            response = execution_result.dict()
            response['rate_limit_remaining'] = remaining
            
            return web.json_response(response)
            
        except Exception as e:
            logger.error(f"Tool execution error: {str(e)}")
            return web.json_response({
                'error': str(e),
                'tool': tool_name
            }, status=500)
    
    async def _execute_tool_impl(
        self, 
        tool_name: str, 
        parameters: Dict[str, Any], 
        agent_id: str
    ) -> Dict[str, Any]:
        """Implementation of tool execution"""
        
        try:
            if tool_name == "job_submit":
                return await self._tool_job_submit(parameters, agent_id)
            elif tool_name == "job_status":
                return await self._tool_job_status(parameters)
            elif tool_name == "job_list":
                return await self._tool_job_list(parameters, agent_id)
            elif tool_name == "credit_balance":
                return await self._tool_credit_balance(agent_id)
            elif tool_name == "credit_transfer":
                return await self._tool_credit_transfer(parameters, agent_id)
            elif tool_name == "credit_history":
                return await self._tool_credit_history(parameters, agent_id)
            elif tool_name == "node_status":
                return await self._tool_node_status(parameters)
            elif tool_name == "node_list":
                return await self._tool_node_list(parameters)
            elif tool_name == "node_contribution":
                return await self._tool_node_contribution(parameters)
            elif tool_name == "wallet_balance":
                return await self._tool_wallet_balance(agent_id)
            elif tool_name == "wallet_transactions":
                return await self._tool_wallet_transactions(parameters, agent_id)
            elif tool_name == "agent_spawn":
                return await self._tool_agent_spawn(parameters, agent_id)
            elif tool_name == "agent_self_terminate":
                return await self._tool_agent_terminate(parameters, agent_id)
            else:
                return {'status': 'error', 'error': f'Unknown tool: {tool_name}'}
                
        except Exception as e:
            return {'status': 'error', 'error': str(e)}
    
    # Tool implementations
    async def _tool_job_submit(self, params: Dict, agent_id: str) -> Dict:
        """Submit a job"""
        spec = params.get('spec', {})
        priority = params.get('priority', 'normal')
        
        # Calculate credit cost based on job spec
        credit_cost = Config.CREDIT_SUBMISSION_COST
        if priority == 'high':
            credit_cost *= 2
        
        job_id = f"job-{uuid.uuid4().hex[:12]}"
        job = JobSubmission(
            job_id=job_id,
            user_id=agent_id,
            spec=spec,
            credit_cost=credit_cost,
            orchestrator="default"
        )
        self.jobs[job_id] = job
        
        # Broadcast job submitted
        await self.ws_manager.broadcast('jobs', {
            'type': 'job_submitted',
            'job_id': job_id,
            'agent_id': agent_id,
            'credit_cost': credit_cost,
            'timestamp': datetime.utcnow().isoformat()
        })
        
        return {
            'status': 'success',
            'data': {
                'job_id': job_id,
                'status': 'pending',
                'credit_cost': credit_cost,
                'message': 'Job submitted successfully'
            }
        }
    
    async def _tool_job_status(self, params: Dict) -> Dict:
        """Get job status"""
        job_id = params.get('job_id')
        
        if not job_id or job_id not in self.jobs:
            return {'status': 'error', 'error': 'Job not found'}
        
        # Simulate job status (in production, query orchestrator)
        import random
        statuses = ['pending', 'running', 'completed']
        job_status = random.choice(statuses)
        
        return {
            'status': 'success',
            'data': {
                'job_id': job_id,
                'status': job_status,
                'timestamp': datetime.utcnow().isoformat()
            }
        }
    
    async def _tool_job_list(self, params: Dict, agent_id: str) -> Dict:
        """List agent's jobs"""
        status_filter = params.get('status', 'all')
        limit = params.get('limit', 20)
        
        agent_jobs = [
            {'job_id': j.job_id, 'status': 'completed', 'credit_cost': j.credit_cost}
            for j in self.jobs.values()
            if j.user_id == agent_id
        ][:limit]
        
        return {
            'status': 'success',
            'data': {
                'jobs': agent_jobs,
                'total': len(agent_jobs)
            }
        }
    
    async def _tool_credit_balance(self, agent_id: str) -> Dict:
        """Get credit balance"""
        # Check if agent exists
        if agent_id in self.agents:
            agent = self.agents[agent_id]
            balance = agent.credits_earned - agent.credits_spent
        elif agent_id in self.users:
            balance = self.users[agent_id].credit_balance
        else:
            balance = 0.0
        
        return {
            'status': 'success',
            'data': {
                'balance': balance,
                'agent_id': agent_id
            }
        }
    
    async def _tool_credit_transfer(self, params: Dict, agent_id: str) -> Dict:
        """Transfer credits"""
        to_id = params.get('to')
        amount = params.get('amount', 0)
        
        if not to_id or amount <= 0:
            return {'status': 'error', 'error': 'Invalid parameters'}
        
        # In production, implement actual transfer
        return {
            'status': 'success',
            'data': {
                'from': agent_id,
                'to': to_id,
                'amount': amount,
                'timestamp': datetime.utcnow().isoformat()
            }
        }
    
    async def _tool_credit_history(self, params: Dict, agent_id: str) -> Dict:
        """Get credit history"""
        limit = params.get('limit', 50)
        
        # Return sample history
        return {
            'status': 'success',
            'data': {
                'transactions': [
                    {
                        'type': 'earn',
                        'amount': 10.5,
                        'description': 'Job execution reward',
                        'timestamp': datetime.utcnow().isoformat()
                    }
                ],
                'agent_id': agent_id
            }
        }
    
    async def _tool_node_status(self, params: Dict) -> Dict:
        """Get node status"""
        node_id = params.get('node_id')
        
        if not node_id or node_id not in self.nodes:
            return {'status': 'error', 'error': 'Node not found'}
        
        node = self.nodes[node_id]
        return {
            'status': 'success',
            'data': {
                'node_id': node_id,
                'status': node.status.value,
                'arch': node.arch.value,
                'credits_earned': node.credits_earned,
                'last_seen': node.last_seen.isoformat()
            }
        }
    
    async def _tool_node_list(self, params: Dict) -> Dict:
        """List nodes"""
        status_filter = params.get('status', 'online')
        limit = params.get('limit', 50)
        
        nodes = []
        for node in self.nodes.values():
            if status_filter != 'all' and node.status.value != status_filter:
                continue
            nodes.append({
                'node_id': node.node_id,
                'status': node.status.value,
                'arch': node.arch.value,
                'credits_earned': node.credits_earned
            })
        
        return {
            'status': 'success',
            'data': {
                'nodes': nodes[:limit],
                'total': len(nodes)
            }
        }
    
    async def _tool_node_contribution(self, params: Dict) -> Dict:
        """Get node contribution stats"""
        node_id = params.get('node_id')
        
        if not node_id or node_id not in self.nodes:
            return {'status': 'error', 'error': 'Node not found'}
        
        node = self.nodes[node_id]
        return {
            'status': 'success',
            'data': {
                'node_id': node_id,
                'cpu_usage_hours': node.cpu_usage_hours,
                'gpu_usage_hours': node.gpu_usage_hours,
                'tier': node.get_tier().value,
                'credits_earned': node.credits_earned
            }
        }
    
    async def _tool_wallet_balance(self, agent_id: str) -> Dict:
        """Get wallet balance"""
        return {
            'status': 'success',
            'data': {
                'agent_id': agent_id,
                'balance': 100.0,  # Sample balance
                'currency': 'DEP',
                'timestamp': datetime.utcnow().isoformat()
            }
        }
    
    async def _tool_wallet_transactions(self, params: Dict, agent_id: str) -> Dict:
        """Get wallet transactions"""
        limit = params.get('limit', 50)
        
        return {
            'status': 'success',
            'data': {
                'transactions': [],
                'agent_id': agent_id,
                'limit': limit
            }
        }
    
    async def _tool_agent_spawn(self, params: Dict, parent_agent_id: str) -> Dict:
        """Spawn a new agent"""
        name = params.get('name', f'Spawned-{uuid.uuid4().hex[:6]}')
        config = params.get('config', {})
        
        new_agent_id = f"agent-{uuid.uuid4().hex[:12]}"
        agent_config = AgentConfig(**config) if config else AgentConfig()
        
        new_agent = Agent(
            agent_id=new_agent_id,
            name=name,
            node_id="virtual",
            status=AgentStatus.INITIALIZING,
            tools=agent_config.tools_enabled,
            last_heartbeat=datetime.utcnow(),
            created_at=datetime.utcnow(),
            config=agent_config
        )
        
        self.agents[new_agent_id] = new_agent
        
        # Broadcast new agent
        await self.ws_manager.broadcast('agents', {
            'type': 'agent_spawned',
            'parent_agent_id': parent_agent_id,
            'agent_id': new_agent_id,
            'name': name,
            'timestamp': datetime.utcnow().isoformat()
        })
        
        return {
            'status': 'success',
            'data': {
                'agent_id': new_agent_id,
                'name': name,
                'message': 'Agent spawned successfully'
            }
        }
    
    async def _tool_agent_terminate(self, params: Dict, agent_id: str) -> Dict:
        """Terminate current agent"""
        confirm = params.get('confirm', False)
        
        if not confirm:
            return {
                'status': 'error',
                'error': 'Termination requires confirm=true'
            }
        
        if agent_id in self.agents:
            agent = self.agents[agent_id]
            agent.status = AgentStatus.OFFLINE
            
            # Broadcast termination
            await self.ws_manager.broadcast('agents', {
                'type': 'agent_terminated',
                'agent_id': agent_id,
                'name': agent.name,
                'timestamp': datetime.utcnow().isoformat()
            })
        
        return {
            'status': 'success',
            'data': {
                'agent_id': agent_id,
                'message': 'Agent terminated'
            }
        }
    
    # ============ WebSocket Handler ============
    
    async def websocket_handler(self, request: web.Request):
        """Handle WebSocket connections for real-time updates"""
        ws = web.WebSocketResponse(heartbeat=30)
        await ws.prepare(request)
        
        client_id = str(uuid.uuid4())
        await self.ws_manager.connect(client_id, ws)
        
        # Get optional agent_id from query
        agent_id = request.query.get('agent_id')
        if agent_id:
            await self.ws_manager.register_agent_ws(agent_id, client_id)
        
        # Subscribe to default channels
        channels = request.query.get('channels', 'jobs,nodes,agents').split(',')
        for channel in channels:
            await self.ws_manager.subscribe(client_id, channel)
        
        try:
            # Send initial connection message
            await ws.send_json({
                'type': 'connected',
                'client_id': client_id,
                'channels': channels,
                'timestamp': datetime.utcnow().isoformat()
            })
            
            # Send current state
            await ws.send_json({
                'type': 'state_sync',
                'data': {
                    'nodes_online': len([n for n in self.nodes.values() if n.status == NodeStatus.ONLINE]),
                    'agents_online': len([a for a in self.agents.values() if a.status != AgentStatus.OFFLINE]),
                    'active_jobs': len(self.jobs)
                }
            })
            
            async for msg in ws:
                if msg.type == WSMsgType.TEXT:
                    try:
                        data = json.loads(msg.data)
                        await self._handle_ws_message(client_id, data)
                    except json.JSONDecodeError:
                        await ws.send_json({'type': 'error', 'error': 'Invalid JSON'})
                
                elif msg.type == WSMsgType.ERROR:
                    logger.error(f"WebSocket error: {ws.exception()}")
                    break
        
        finally:
            await self.ws_manager.disconnect(client_id)
        
        return ws
    
    async def _handle_ws_message(self, client_id: str, data: Dict[str, Any]):
        """Handle incoming WebSocket message"""
        msg_type = data.get('type')
        
        if msg_type == 'subscribe':
            channel = data.get('channel')
            if channel:
                await self.ws_manager.subscribe(client_id, channel)
                await self.ws_manager.send_to_client(client_id, {
                    'type': 'subscribed',
                    'channel': channel
                })
        
        elif msg_type == 'unsubscribe':
            channel = data.get('channel')
            if channel:
                await self.ws_manager.unsubscribe(client_id, channel)
                await self.ws_manager.send_to_client(client_id, {
                    'type': 'unsubscribed',
                    'channel': channel
                })
        
        elif msg_type == 'ping':
            await self.ws_manager.send_to_client(client_id, {
                'type': 'pong',
                'timestamp': datetime.utcnow().isoformat()
            })
        
        elif msg_type == 'agent_command':
            # Forward command to specific agent
            target_agent_id = data.get('agent_id')
            command = data.get('command')
            if target_agent_id and command:
                await self.ws_manager.send_to_agent(target_agent_id, {
                    'type': 'command',
                    'command': command,
                    'from_client': client_id,
                    'timestamp': datetime.utcnow().isoformat()
                })
    
    async def start(self):
        """Start the bootstrap server"""
        # Initialize database connections
        # self.db_pool = await asyncpg.create_pool(Config.DATABASE_URL)
        # self.redis_client = redis.from_url(Config.REDIS_URL)
        
        # Create aiohttp session
        self.session = aiohttp.ClientSession()
        
        # Start background tasks
        asyncio.create_task(self._cleanup_task())
        asyncio.create_task(self._metrics_task())
        
        # Start web server
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, Config.HOST, Config.PORT)
        
        logger.info(f"DEparrow Bootstrap Server starting on {Config.HOST}:{Config.PORT}")
        await site.start()
        
        # Keep running
        await asyncio.Event().wait()
    
    async def _cleanup_task(self):
        """Background task to cleanup stale nodes and agents"""
        while True:
            try:
                now = datetime.utcnow()
                stale_threshold = now - timedelta(seconds=Config.NODE_REGISTRATION_TTL)
                
                # Cleanup nodes
                for node_id, node in list(self.nodes.items()):
                    if node.last_seen < stale_threshold:
                        node.status = NodeStatus.OFFLINE
                        logger.info(f"Marked node as offline: {node_id}")
                
                # Cleanup agents (longer timeout for agents - 10 minutes)
                agent_stale_threshold = now - timedelta(seconds=600)
                for agent_id, agent in list(self.agents.items()):
                    if agent.last_heartbeat < agent_stale_threshold:
                        agent.status = AgentStatus.OFFLINE
                        logger.info(f"Marked agent as offline: {agent_id}")
                        
                        # Broadcast offline status
                        await self.ws_manager.broadcast('agents', {
                            'type': 'agent_offline',
                            'agent_id': agent_id,
                            'timestamp': now.isoformat()
                        })
                
            except Exception as e:
                logger.error(f"Cleanup task error: {str(e)}")
            
            await asyncio.sleep(60)  # Run every minute
    
    async def _metrics_task(self):
        """Background task to update metrics and broadcast updates"""
        while True:
            try:
                # Update orchestrator node counts
                for orchestrator in self.orchestrators.values():
                    orchestrator.node_count = len([
                        n for n in self.nodes.values() 
                        if n.status == NodeStatus.ONLINE
                    ]) // max(len(self.orchestrators), 1)
                
                # Broadcast periodic metrics update
                metrics = {
                    'type': 'metrics_update',
                    'data': {
                        'nodes_online': len([n for n in self.nodes.values() if n.status == NodeStatus.ONLINE]),
                        'agents_online': len([a for a in self.agents.values() if a.status != AgentStatus.OFFLINE]),
                        'active_jobs': len(self.jobs),
                        'ws_connections': self.ws_manager.get_connection_count()
                    },
                    'timestamp': datetime.utcnow().isoformat()
                }
                await self.ws_manager.broadcast('metrics', metrics)
                
            except Exception as e:
                logger.error(f"Metrics task error: {str(e)}")
            
            await asyncio.sleep(30)  # Run every 30 seconds
    
    async def stop(self):
        """Stop the bootstrap server"""
        if self.session:
            await self.session.close()
        if self.db_pool:
            await self.db_pool.close()
        if self.redis_client:
            await self.redis_client.close()

def main():
    """Main entry point"""
    server = DEparrowBootstrapServer()
    
    try:
        asyncio.run(server.start())
    except KeyboardInterrupt:
        logger.info("Shutting down DEparrow Bootstrap Server...")
    except Exception as e:
        logger.error(f"Server error: {str(e)}")
        sys.exit(1)
    finally:
        asyncio.run(server.stop())

if __name__ == "__main__":
    main()
#!/usr/bin/env python3
"""
DEparrow Bootstrap Server
Meta-OS Control Plane - Replaces default Bacalhau bootstrap
"""

import asyncio
import json
import logging
import os
import sys
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, asdict
from enum import Enum

import aiohttp
from aiohttp import web
import jwt
import redis.asyncio as redis
from pydantic import BaseModel, Field, validator
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

class DEparrowBootstrapServer:
    """DEparrow Bootstrap Server - Meta-OS Control Plane"""
    
    def __init__(self):
        self.app = web.Application()
        self.setup_routes()
        self.db_pool = None
        self.redis_client = None
        self.session = None
        
        # In-memory caches
        self.nodes: Dict[str, Node] = {}
        self.orchestrators: Dict[str, Orchestrator] = {}
        self.users: Dict[str, User] = {}
        self.jobs: Dict[str, JobSubmission] = {}
        
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
    
    def setup_routes(self):
        """Setup API routes"""
        self.app.router.add_post('/api/v1/nodes/register', self.register_node)
        self.app.router.add_get('/api/v1/nodes', self.list_nodes)
        self.app.router.add_get('/api/v1/nodes/{node_id}', self.get_node)
        self.app.router.add_post('/api/v1/nodes/{node_id}/heartbeat', self.node_heartbeat)
        
        # Contribution tracking routes
        self.app.router.add_get('/api/v1/nodes/{node_id}/contribution', self.get_node_contribution)
        self.app.router.add_get('/api/v1/network/contribution', self.get_network_contribution)
        self.app.router.add_get('/api/v1/network/leaderboard', self.get_leaderboard)
        self.app.router.add_get('/api/v1/network/globe', self.get_globe_data)
        
        self.app.router.add_post('/api/v1/orchestrators/register', self.register_orchestrator)
        self.app.router.add_get('/api/v1/orchestrators', self.list_orchestrators)
        
        self.app.router.add_post('/api/v1/jobs/submit', self.submit_job)
        self.app.router.add_get('/api/v1/jobs/{job_id}', self.get_job_status)
        self.app.router.add_post('/api/v1/jobs/{job_id}/cancel', self.cancel_job)
        
        self.app.router.add_post('/api/v1/credits/check', self.check_credits)
        self.app.router.add_post('/api/v1/credits/transfer', self.transfer_credits)
        self.app.router.add_get('/api/v1/credits/balance/{user_id}', self.get_credit_balance)
        
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
            'version': '1.0.0',
            'components': {
                'nodes': len(self.nodes),
                'orchestrators': len(self.orchestrators),
                'users': len(self.users),
                'jobs': len(self.jobs)
            }
        })
    
    async def get_metrics(self, request: web.Request):
        """Get system metrics"""
        online_nodes = len([n for n in self.nodes.values() if n.status == NodeStatus.ONLINE])
        total_credits = sum(user.credit_balance for user in self.users.values())
        node_credits = sum(node.credits_earned for node in self.nodes.values())
        
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
                    'active': len([j for j in self.jobs.values()])  # Would have status in production
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
        """Background task to cleanup stale nodes"""
        while True:
            try:
                now = datetime.utcnow()
                stale_threshold = now - timedelta(seconds=Config.NODE_REGISTRATION_TTL)
                
                for node_id, node in list(self.nodes.items()):
                    if node.last_seen < stale_threshold:
                        node.status = NodeStatus.OFFLINE
                        logger.info(f"Marked node as offline: {node_id}")
                
            except Exception as e:
                logger.error(f"Cleanup task error: {str(e)}")
            
            await asyncio.sleep(60)  # Run every minute
    
    async def _metrics_task(self):
        """Background task to update metrics"""
        while True:
            try:
                # Update orchestrator node counts
                for orchestrator in self.orchestrators.values():
                    # Count nodes connected to this orchestrator
                    # In production, would query orchestrator
                    orchestrator.node_count = len([
                        n for n in self.nodes.values() 
                        if n.status == NodeStatus.ONLINE
                    ]) // max(len(self.orchestrators), 1)
                
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
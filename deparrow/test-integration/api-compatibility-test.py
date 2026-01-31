#!/usr/bin/env python3
"""
DEparrow API Compatibility Test Suite
Tests all API endpoints and data contracts between layers
"""

import json
import os
import sys
import time
import asyncio
import aiohttp
import subprocess
from pathlib import Path
from typing import Dict, Any, List, Optional

class DEparrowAPITester:
    def __init__(self):
        self.bootstrap_url = os.getenv('DEPARROW_BOOTSTRAP', 'http://localhost:8080')
        self.api_key = os.getenv('DEPARROW_API_KEY', 'test-api-key-12345')
        self.test_results = []
        self.session = None
        
    async def __aenter__(self):
        self.session = aiohttp.ClientSession(
            headers={'Authorization': f'Bearer {self.api_key}'},
            timeout=aiohttp.ClientTimeout(total=30)
        )
        return self
        
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
    
    async def test_health_endpoint(self) -> bool:
        """Test the health check endpoint"""
        try:
            async with self.session.get(f'{self.bootstrap_url}/api/v1/health') as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if data.get('status') == 'healthy':
                        self.log_test('Health Endpoint', True, 'System healthy')
                        return True
                    else:
                        self.log_test('Health Endpoint', False, f'Unhealthy status: {data}')
                        return False
                else:
                    self.log_test('Health Endpoint', False, f'HTTP {resp.status}')
                    return False
        except Exception as e:
            self.log_test('Health Endpoint', False, str(e))
            return False
    
    async def test_node_registration(self) -> Optional[str]:
        """Test node registration and return node ID"""
        node_data = {
            'node_id': f'test-node-{int(time.time())}',
            'public_key': 'test-public-key-12345',
            'resources': {
                'cpu': 4,
                'memory': '4GB',
                'disk': '20GB',
                'arch': 'x86_64'
            }
        }
        
        try:
            async with self.session.post(
                f'{self.bootstrap_url}/api/v1/nodes/register',
                json=node_data
            ) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if data.get('success') and data.get('node_id'):
                        self.log_test('Node Registration', True, f"Node {data['node_id']} registered")
                        return data['node_id']
                    else:
                        self.log_test('Node Registration', False, f'Registration failed: {data}')
                        return None
                else:
                    self.log_test('Node Registration', False, f'HTTP {resp.status}')
                    return None
        except Exception as e:
            self.log_test('Node Registration', False, str(e))
            return None
    
    async def test_job_submission(self, node_id: str) -> Optional[str]:
        """Test job submission and return job ID"""
        job_data = {
            'node_id': node_id,
            'job_spec': {
                'engine': 'docker',
                'image': 'python:3.9-slim',
                'command': ['echo', 'Hello DEparrow'],
                'resources': {'cpu': 1, 'memory': '512MB'}
            },
            'credits': 10
        }
        
        try:
            async with self.session.post(
                f'{self.bootstrap_url}/api/v1/jobs/submit',
                json=job_data
            ) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if data.get('job_id'):
                        self.log_test('Job Submission', True, f"Job {data['job_id']} submitted")
                        return data['job_id']
                    else:
                        self.log_test('Job Submission', False, f'No job ID returned: {data}')
                        return None
                else:
                    self.log_test('Job Submission', False, f'HTTP {resp.status}')
                    return None
        except Exception as e:
            self.log_test('Job Submission', False, str(e))
            return None
    
    async def test_insufficient_credits(self, node_id: str) -> bool:
        """Test job rejection with insufficient credits"""
        job_data = {
            'node_id': node_id,
            'job_spec': {'engine': 'docker', 'image': 'python:3.9'},
            'credits': 5  # Below minimum requirement
        }
        
        try:
            async with self.session.post(
                f'{self.bootstrap_url}/api/v1/jobs/submit',
                json=job_data
            ) as resp:
                # Should return error for insufficient credits
                if resp.status == 400:
                    data = await resp.json()
                    if 'error' in data and 'credit' in data.get('error', '').lower():
                        self.log_test('Credit Validation', True, 'Correctly rejected low credits')
                        return True
                    else:
                        self.log_test('Credit Validation', False, f'Wrong error: {data}')
                        return False
                else:
                    self.log_test('Credit Validation', False, f'Expected 400, got {resp.status}')
                    return False
        except Exception as e:
            self.log_test('Credit Validation', False, str(e))
            return False
    
    async def test_credit_transfer(self) -> bool:
        """Test credit transfer between users"""
        transfer_data = {
            'from_user': 'user1',
            'to_user': 'user2',
            'amount': 50,
            'reason': 'Test transfer'
        }
        
        try:
            async with self.session.post(
                f'{self.bootstrap_url}/api/v1/credits/transfer',
                json=transfer_data
            ) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if data.get('success') and data.get('transaction_id'):
                        self.log_test('Credit Transfer', True, f"Transaction {data['transaction_id']}")
                        return True
                    else:
                        self.log_test('Credit Transfer', False, f'Transfer failed: {data}')
                        return False
                else:
                    self.log_test('Credit Transfer', False, f'HTTP {resp.status}')
                    return False
        except Exception as e:
            self.log_test('Credit Transfer', False, str(e))
            return False
    
    async def test_credit_check(self) -> bool:
        """Test credit balance checking"""
        try:
            async with self.session.get(f'{self.bootstrap_url}/api/v1/credits/balance') as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if 'balance' in data:
                        self.log_test('Credit Check', True, f"Balance: {data['balance']} credits")
                        return True
                    else:
                        self.log_test('Credit Check', False, f'No balance in response: {data}')
                        return False
                else:
                    self.log_test('Credit Check', False, f'HTTP {resp.status}')
                    return False
        except Exception as e:
            self.log_test('Credit Check', False, str(e))
            return False
    
    async def test_node_status(self, node_id: str) -> bool:
        """Test node status checking"""
        try:
            async with self.session.get(f'{self.bootstrap_url}/api/v1/nodes/{node_id}/status') as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if 'status' in data:
                        self.log_test('Node Status', True, f"Status: {data['status']}")
                        return True
                    else:
                        self.log_test('Node Status', False, f'No status in response: {data}')
                        return False
                else:
                    self.log_test('Node Status', False, f'HTTP {resp.status}')
                    return False
        except Exception as e:
            self.log_test('Node Status', False, str(e))
            return False
    
    def log_test(self, test_name: str, success: bool, message: str):
        """Log test result"""
        status = "✅ PASS" if success else "❌ FAIL"
        result = {
            'test': test_name,
            'success': success,
            'message': message,
            'timestamp': time.time()
        }
        self.test_results.append(result)
        
        print(f"{status} {test_name}: {message}")
    
    def validate_json_schema(self, data: Dict[str, Any], required_fields: List[str]) -> bool:
        """Validate JSON response has required fields"""
        missing_fields = [field for field in required_fields if field not in data]
        if missing_fields:
            return False
        return True
    
    def test_file_structure(self) -> bool:
        """Test that all required files exist"""
        project_root = Path(__file__).parent.parent.parent
        
        required_files = [
            'alpine-layer/Dockerfile',
            'alpine-layer/scripts/init-node.sh',
            'alpine-layer/scripts/health-check.sh',
            'metaos-layer/bootstrap-server.py',
            'gui-layer/src/pages/Dashboard.tsx',
            'gui-layer/src/api/client.ts',
            'gui-layer/src/contexts/AuthContext.tsx',
        ]
        
        all_exist = True
        for file_path in required_files:
            full_path = project_root / file_path
            if full_path.exists():
                print(f"✅ Found: {file_path}")
            else:
                print(f"❌ Missing: {file_path}")
                all_exist = False
        
        return all_exist
    
    async def run_all_tests(self) -> Dict[str, Any]:
        """Run complete API test suite"""
        print("🔍 Testing DEparrow API Compatibility...")
        print("=" * 50)
        
        # Test file structure first
        print("\n📁 Testing File Structure...")
        file_structure_ok = self.test_file_structure()
        
        if not file_structure_ok:
            self.log_test('File Structure', False, 'Missing required files')
            return self.generate_report()
        
        # Test API endpoints
        print("\n🌐 Testing API Endpoints...")
        
        # Health check
        await self.test_health_endpoint()
        
        # Node operations
        node_id = await self.test_node_registration()
        if node_id:
            await self.test_node_status(node_id)
            await self.test_insufficient_credits(node_id)
        
        # Job operations
        if node_id:
            job_id = await self.test_job_submission(node_id)
            if job_id:
                # Test job status if endpoint exists
                try:
                    async with self.session.get(f'{self.bootstrap_url}/api/v1/jobs/{job_id}/status') as resp:
                        if resp.status == 200:
                            self.log_test('Job Status', True, f"Job {job_id} status retrieved")
                        else:
                            self.log_test('Job Status', False, f'HTTP {resp.status}')
                except:
                    self.log_test('Job Status', False, 'Job status endpoint not available')
        
        # Credit operations
        await self.test_credit_check()
        await self.test_credit_transfer()
        
        return self.generate_report()
    
    def generate_report(self) -> Dict[str, Any]:
        """Generate test report"""
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results if result['success'])
        failed_tests = total_tests - passed_tests
        
        report = {
            'total_tests': total_tests,
            'passed': passed_tests,
            'failed': failed_tests,
            'success_rate': (passed_tests / total_tests * 100) if total_tests > 0 else 0,
            'test_results': self.test_results,
            'timestamp': time.time()
        }
        
        print("\n" + "=" * 50)
        print(f"📊 Test Results: {passed_tests}/{total_tests} passed ({report['success_rate']:.1f}%)")
        
        if failed_tests > 0:
            print("\n❌ Failed Tests:")
            for result in self.test_results:
                if not result['success']:
                    print(f"  • {result['test']}: {result['message']}")
        
        return report

async def main():
    """Main test execution"""
    print("🚀 DEparrow API Compatibility Test Suite")
    print("=" * 50)
    
    # Create test instance
    async with DEparrowAPITester() as tester:
        report = await tester.run_all_tests()
        
        # Save report to file
        report_file = Path(__file__).parent / 'api-test-report.json'
        with open(report_file, 'w') as f:
            json.dump(report, f, indent=2)
        
        print(f"\n📄 Detailed report saved to: {report_file}")
        
        # Return exit code based on results
        if report['failed'] > 0:
            print("\n❌ Some tests failed!")
            return 1
        else:
            print("\n🎉 All API compatibility tests passed!")
            return 0

if __name__ == '__main__':
    try:
        exit_code = asyncio.run(main())
        sys.exit(exit_code)
    except KeyboardInterrupt:
        print("\n⚠️ Test interrupted by user")
        sys.exit(130)
    except Exception as e:
        print(f"\n💥 Test suite error: {e}")
        sys.exit(1)

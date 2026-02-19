"""Shared pytest fixtures for bacalhau-sdk tests."""

import pytest
from unittest.mock import MagicMock, patch
from pathlib import Path
import tempfile
import os


@pytest.fixture
def mock_config():
    """Create a mock Configuration object."""
    from bacalhau_apiclient import Configuration

    config = Configuration()
    config.host = "http://localhost:1234"
    return config


@pytest.fixture
def mock_api_client():
    """Create a mock API client."""
    mock_client = MagicMock()
    return mock_client


@pytest.fixture
def mock_orchestrator_api():
    """Create a mock OrchestratorApi instance."""
    mock_api = MagicMock()
    return mock_api


@pytest.fixture
def temp_config_dir():
    """Create a temporary directory for config testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        old_dir = os.environ.get("BACALHAU_DIR", "")
        os.environ["BACALHAU_DIR"] = tmpdir
        yield Path(tmpdir)
        if old_dir:
            os.environ["BACALHAU_DIR"] = old_dir
        else:
            os.environ.pop("BACALHAU_DIR", None)


@pytest.fixture
def clean_env():
    """Clean environment variables for testing."""
    env_vars = [
        "BACALHAU_DIR",
        "BACALHAU_HTTPS",
        "BACALHAU_API_HOST",
        "BACALHAU_API_PORT",
    ]
    original_values = {}
    for var in env_vars:
        original_values[var] = os.environ.get(var)
        if var in os.environ:
            del os.environ[var]

    yield

    for var, value in original_values.items():
        if value is not None:
            os.environ[var] = value
        elif var in os.environ:
            del os.environ[var]


@pytest.fixture
def sample_job_id():
    """Return a sample job ID for testing."""
    return "job-123-abc-def-456"


@pytest.fixture
def sample_put_job_request():
    """Create a sample PutJobRequest for testing."""
    from bacalhau_apiclient.models.api_put_job_request import ApiPutJobRequest

    request = MagicMock(spec=ApiPutJobRequest)
    return request


@pytest.fixture
def sample_put_job_response():
    """Create a sample PutJobResponse for testing."""
    from bacalhau_apiclient.models.api_put_job_response import ApiPutJobResponse

    response = MagicMock(spec=ApiPutJobResponse)
    response.job = MagicMock()
    response.job.id = "job-123-abc-def-456"
    return response


@pytest.fixture
def sample_get_job_response():
    """Create a sample GetJobResponse for testing."""
    from bacalhau_apiclient.models.api_get_job_response import ApiGetJobResponse

    response = MagicMock(spec=ApiGetJobResponse)
    response.job = MagicMock()
    response.job.id = "job-123-abc-def-456"
    response.job.state = MagicMock()
    return response


@pytest.fixture
def sample_list_jobs_response():
    """Create a sample ListJobsResponse for testing."""
    from bacalhau_apiclient.models.api_list_jobs_response import ApiListJobsResponse

    response = MagicMock(spec=ApiListJobsResponse)
    response.jobs = []
    return response


@pytest.fixture
def sample_stop_job_response():
    """Create a sample StopJobResponse for testing."""
    from bacalhau_apiclient.models.api_stop_job_response import ApiStopJobResponse

    response = MagicMock(spec=ApiStopJobResponse)
    return response


@pytest.fixture
def sample_job_executions_response():
    """Create a sample ListJobExecutionsResponse for testing."""
    from bacalhau_apiclient.models.api_list_job_executions_response import (
        ApiListJobExecutionsResponse,
    )

    response = MagicMock(spec=ApiListJobExecutionsResponse)
    response.executions = []
    return response


@pytest.fixture
def sample_job_results_response():
    """Create a sample ListJobResultsResponse for testing."""
    from bacalhau_apiclient.models.api_list_job_results_response import (
        ApiListJobResultsResponse,
    )

    response = MagicMock(spec=ApiListJobResultsResponse)
    response.results = []
    return response


@pytest.fixture
def sample_job_history_response():
    """Create a sample ListJobHistoryResponse for testing."""
    from bacalhau_apiclient.models.api_list_job_history_response import (
        ApiListJobHistoryResponse,
    )

    response = MagicMock(spec=ApiListJobHistoryResponse)
    response.history = []
    return response

"""Tests for Jobs class."""

import pytest
from unittest.mock import MagicMock, patch

from bacalhau_sdk.jobs import Jobs
from bacalhau_apiclient.rest import ApiException


class TestJobsInit:
    """Tests for Jobs class initialization."""

    def test_init_creates_orchestrator_service(self):
        """Test that Jobs.__init__ creates an OrchestratorService."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            mock_service_class.return_value = mock_service

            jobs = Jobs()

            mock_service_class.assert_called_once()

    def test_init_with_mocked_config(self, mock_config):
        """Test initialization with mocked configuration."""
        with patch("bacalhau_sdk.jobs.init_config") as mock_init_config:
            mock_init_config.return_value = mock_config

            with patch(
                "bacalhau_sdk.jobs.OrchestratorService"
            ) as mock_service_class:
                jobs = Jobs()

                mock_init_config.assert_called_once()
                mock_service_class.assert_called_once_with(config=mock_config)


class TestJobsPut:
    """Tests for Jobs.put method."""

    def test_put_calls_orchestrator_put_job(self, sample_put_job_request):
        """Test that put calls orchestrator_service.put_job."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.put_job.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.put(sample_put_job_request)

            mock_service.put_job.assert_called_once_with(sample_put_job_request)
            assert result == expected_response

    def test_put_returns_put_job_response(self, sample_put_job_request):
        """Test that put returns the expected response type."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            mock_response = MagicMock()
            mock_response.job = MagicMock()
            mock_response.job.id = "test-job-id"
            mock_service.put_job.return_value = mock_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.put(sample_put_job_request)

            assert hasattr(result, "job")
            assert result.job.id == "test-job-id"


class TestJobsStop:
    """Tests for Jobs.stop method."""

    def test_stop_calls_orchestrator_stop_job(self, sample_job_id):
        """Test that stop calls orchestrator_service.stop_job."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.stop_job.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.stop(job_id=sample_job_id, reason="User requested")

            mock_service.stop_job.assert_called_once_with(
                id=sample_job_id, reason="User requested"
            )
            assert result == expected_response

    def test_stop_without_reason(self, sample_job_id):
        """Test stop without a reason parameter."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.stop_job.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.stop(job_id=sample_job_id)

            mock_service.stop_job.assert_called_once_with(
                id=sample_job_id, reason=None
            )
            assert result == expected_response


class TestJobsExecutions:
    """Tests for Jobs.executions method."""

    def test_executions_calls_orchestrator_with_defaults(self, sample_job_id):
        """Test executions with default parameters."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.job_executions.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.executions(job_id=sample_job_id)

            mock_service.job_executions.assert_called_once_with(
                id=sample_job_id,
                namespace="",
                limit=5,
                next_token="",
                reverse=False,
                order_by="",
            )
            assert result == expected_response

    def test_executions_with_all_params(self, sample_job_id):
        """Test executions with all parameters specified."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.job_executions.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.executions(
                job_id=sample_job_id,
                namespace="production",
                next_token="next-token-123",
                limit=50,
                reverse=True,
                order_by="completed_at",
            )

            mock_service.job_executions.assert_called_once_with(
                id=sample_job_id,
                namespace="production",
                limit=50,
                next_token="next-token-123",
                reverse=True,
                order_by="completed_at",
            )
            assert result == expected_response


class TestJobsResults:
    """Tests for Jobs.results method."""

    def test_results_calls_orchestrator_job_results(self, sample_job_id):
        """Test that results calls orchestrator_service.job_results."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            expected_response.results = []
            mock_service.job_results.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.results(job_id=sample_job_id)

            mock_service.job_results.assert_called_once_with(id=sample_job_id)
            assert result == expected_response


class TestJobsGet:
    """Tests for Jobs.get method."""

    def test_get_calls_orchestrator_with_defaults(self, sample_job_id):
        """Test get with default parameters."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.get_job.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.get(job_id=sample_job_id)

            mock_service.get_job.assert_called_once_with(
                id=sample_job_id, include="", limit=10
            )
            assert result == expected_response

    def test_get_with_include_and_limit(self, sample_job_id):
        """Test get with include and limit parameters."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.get_job.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.get(job_id=sample_job_id, include="history", limit=25)

            mock_service.get_job.assert_called_once_with(
                id=sample_job_id, include="history", limit=25
            )
            assert result == expected_response


class TestJobsHistory:
    """Tests for Jobs.history method."""

    def test_history_calls_orchestrator_with_defaults(self, sample_job_id):
        """Test history with default parameters."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.job_history.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.history(job_id=sample_job_id)

            mock_service.job_history.assert_called_once_with(
                id=sample_job_id,
                event_type="execution",
                node_id="",
                execution_id="",
            )
            assert result == expected_response

    def test_history_with_all_params(self, sample_job_id):
        """Test history with all parameters specified."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            mock_service.job_history.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.history(
                job_id=sample_job_id,
                event_type="job",
                node_id="node-789",
                execution_id="exec-123",
            )

            mock_service.job_history.assert_called_once_with(
                id=sample_job_id,
                event_type="job",
                node_id="node-789",
                execution_id="exec-123",
            )
            assert result == expected_response


class TestJobsList:
    """Tests for Jobs.list method."""

    def test_list_calls_orchestrator_with_defaults(self):
        """Test list with default parameters."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            expected_response.jobs = []
            mock_service.list_jobs.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.list()

            mock_service.list_jobs.assert_called_once_with(
                limit=5, next_token="", order_by="created_at", reverse=False
            )
            assert result == expected_response

    def test_list_with_all_params(self):
        """Test list with all parameters specified."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            expected_response = MagicMock()
            expected_response.jobs = []
            mock_service.list_jobs.return_value = expected_response
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.list(
                limit=100,
                next_token="pagination-token",
                order_by="updated_at",
                reverse=True,
            )

            mock_service.list_jobs.assert_called_once_with(
                limit=100,
                next_token="pagination-token",
                order_by="updated_at",
                reverse=True,
            )
            assert result == expected_response


class TestJobsIntegration:
    """Integration-style tests for Jobs class workflow."""

    def test_full_job_workflow(self, sample_put_job_request, sample_job_id):
        """Test a typical job workflow: submit, get, list, stop."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            mock_service_class.return_value = mock_service

            # Setup responses
            put_response = MagicMock()
            put_response.job.id = sample_job_id
            mock_service.put_job.return_value = put_response

            get_response = MagicMock()
            get_response.job.id = sample_job_id
            get_response.job.state = "Running"
            mock_service.get_job.return_value = get_response

            list_response = MagicMock()
            list_response.jobs = [get_response.job]
            mock_service.list_jobs.return_value = list_response

            stop_response = MagicMock()
            mock_service.stop_job.return_value = stop_response

            # Execute workflow
            jobs = Jobs()

            # Submit job
            submitted = jobs.put(sample_put_job_request)
            assert submitted.job.id == sample_job_id

            # Get job status
            status = jobs.get(job_id=sample_job_id)
            assert status.job.state == "Running"

            # List jobs
            all_jobs = jobs.list()
            assert len(all_jobs.jobs) == 1

            # Stop job
            stopped = jobs.stop(job_id=sample_job_id, reason="Completed")
            assert stopped is not None

    def test_error_propagation(self, sample_job_id):
        """Test that errors from orchestrator_service are properly propagated."""
        with patch("bacalhau_sdk.jobs.OrchestratorService") as mock_service_class:
            mock_service = MagicMock()
            mock_service.get_job.return_value = None  # Simulates error
            mock_service_class.return_value = mock_service

            jobs = Jobs()
            result = jobs.get(job_id=sample_job_id)

            # When orchestrator_service returns None (error case),
            # the Jobs method should also return None
            assert result is None

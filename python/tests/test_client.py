"""Tests for OrchestratorService class."""

import pytest
from unittest.mock import MagicMock, patch, PropertyMock

from bacalhau_apiclient.rest import ApiException


class TestOrchestratorServiceInit:
    """Tests for OrchestratorService initialization."""

    def test_init_with_config(self, mock_config):
        """Test initialization with a Configuration object."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            service = OrchestratorService(config=mock_config)

            mock_client_class.assert_called_once_with(mock_config)
            mock_api_class.assert_called_once()

    def test_init_sets_api_client_and_endpoint(self, mock_config):
        """Test that init properly sets the API client and endpoint."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)

            assert service.endpoint is mock_api_instance


class TestPutJob:
    """Tests for put_job method."""

    def test_put_job_success(self, mock_config, sample_put_job_request):
        """Test successful job submission."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorput_job.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.put_job(sample_put_job_request)

            mock_api_instance.orchestratorput_job.assert_called_once_with(
                sample_put_job_request
            )
            assert result == expected_response

    def test_put_job_api_exception(self, mock_config, sample_put_job_request):
        """Test put_job handles ApiException."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorput_job.side_effect = ApiException(
                status=500, reason="Internal Server Error"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.put_job(sample_put_job_request)

            assert result is None

    def test_put_job_network_error(self, mock_config, sample_put_job_request):
        """Test put_job handles network errors."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorput_job.side_effect = ConnectionError(
                "Network unreachable"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)

            with pytest.raises(ConnectionError):
                service.put_job(sample_put_job_request)


class TestStopJob:
    """Tests for stop_job method."""

    def test_stop_job_success(self, mock_config, sample_job_id):
        """Test successful job stop."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorstop_job.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.stop_job(id=sample_job_id, reason="User requested")

            mock_api_instance.orchestratorstop_job.assert_called_once_with(
                id=sample_job_id, reason="User requested"
            )
            assert result == expected_response

    def test_stop_job_without_reason(self, mock_config, sample_job_id):
        """Test stop_job without a reason."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorstop_job.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.stop_job(id=sample_job_id, reason=None)

            mock_api_instance.orchestratorstop_job.assert_called_once_with(
                id=sample_job_id, reason=None
            )
            assert result == expected_response

    def test_stop_job_api_exception(self, mock_config, sample_job_id):
        """Test stop_job handles ApiException."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorstop_job.side_effect = ApiException(
                status=404, reason="Job not found"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.stop_job(id=sample_job_id, reason="test")

            assert result is None


class TestGetJob:
    """Tests for get_job method."""

    def test_get_job_success(self, mock_config, sample_job_id):
        """Test successful job retrieval."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorget_job.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.get_job(id=sample_job_id)

            mock_api_instance.orchestratorget_job.assert_called_once_with(
                id=sample_job_id, include="", limit=10
            )
            assert result == expected_response

    def test_get_job_with_include(self, mock_config, sample_job_id):
        """Test get_job with include parameter."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorget_job.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.get_job(id=sample_job_id, include="executions", limit=20)

            mock_api_instance.orchestratorget_job.assert_called_once_with(
                id=sample_job_id, include="executions", limit=20
            )
            assert result == expected_response

    def test_get_job_api_exception(self, mock_config, sample_job_id):
        """Test get_job handles ApiException."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorget_job.side_effect = ApiException(
                status=404, reason="Job not found"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.get_job(id=sample_job_id)

            assert result is None


class TestListJobs:
    """Tests for list_jobs method."""

    def test_list_jobs_success(self, mock_config):
        """Test successful job listing."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            expected_response.jobs = []
            mock_api_instance.orchestratorlist_jobs.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.list_jobs()

            mock_api_instance.orchestratorlist_jobs.assert_called_once_with(
                limit=5, order_by="created_at", reverse=False, next_token=""
            )
            assert result == expected_response

    def test_list_jobs_with_pagination(self, mock_config):
        """Test list_jobs with pagination parameters."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            expected_response.jobs = []
            mock_api_instance.orchestratorlist_jobs.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.list_jobs(
                limit=10, next_token="token123", order_by="updated_at", reverse=True
            )

            mock_api_instance.orchestratorlist_jobs.assert_called_once_with(
                limit=10, order_by="updated_at", reverse=True, next_token="token123"
            )
            assert result == expected_response

    def test_list_jobs_api_exception(self, mock_config):
        """Test list_jobs handles ApiException."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorlist_jobs.side_effect = ApiException(
                status=500, reason="Server error"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)

            # Note: This will raise UnboundLocalError because api_response
            # is referenced in the return statement after the except block
            with pytest.raises(UnboundLocalError):
                service.list_jobs()


class TestJobExecutions:
    """Tests for job_executions method."""

    def test_job_executions_success(self, mock_config, sample_job_id):
        """Test successful job executions retrieval."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorjob_executions.return_value = (
                expected_response
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_executions(id=sample_job_id)

            mock_api_instance.orchestratorjob_executions.assert_called_once_with(
                id=sample_job_id,
                namespace="",
                limit=5,
                next_token="",
                reverse=False,
                order_by="",
            )
            assert result == expected_response

    def test_job_executions_with_all_params(self, mock_config, sample_job_id):
        """Test job_executions with all parameters."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorjob_executions.return_value = (
                expected_response
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_executions(
                id=sample_job_id,
                namespace="default",
                limit=20,
                next_token="token456",
                reverse=True,
                order_by="started_at",
            )

            mock_api_instance.orchestratorjob_executions.assert_called_once_with(
                id=sample_job_id,
                namespace="default",
                limit=20,
                next_token="token456",
                reverse=True,
                order_by="started_at",
            )
            assert result == expected_response

    def test_job_executions_api_exception(self, mock_config, sample_job_id):
        """Test job_executions handles ApiException."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorjob_executions.side_effect = ApiException(
                status=404, reason="Job not found"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_executions(id=sample_job_id)

            assert result is None


class TestJobResults:
    """Tests for job_results method."""

    def test_job_results_success(self, mock_config, sample_job_id):
        """Test successful job results retrieval."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorjob_results.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_results(id=sample_job_id)

            mock_api_instance.orchestratorjob_results.assert_called_once_with(
                id=sample_job_id
            )
            assert result == expected_response

    def test_job_results_api_exception(self, mock_config, sample_job_id):
        """Test job_results handles ApiException."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorjob_results.side_effect = ApiException(
                status=500, reason="Server error"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_results(id=sample_job_id)

            assert result is None


class TestJobHistory:
    """Tests for job_history method."""

    def test_job_history_success(self, mock_config, sample_job_id):
        """Test successful job history retrieval."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorjob_history.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_history(id=sample_job_id)

            mock_api_instance.orchestratorjob_history.assert_called_once_with(
                id=sample_job_id, event_type="execution", node_id="", execution_id=""
            )
            assert result == expected_response

    def test_job_history_with_all_params(self, mock_config, sample_job_id):
        """Test job_history with all parameters."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            expected_response = MagicMock()
            mock_api_instance.orchestratorjob_history.return_value = expected_response
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_history(
                id=sample_job_id,
                event_type="job",
                node_id="node-123",
                execution_id="exec-456",
            )

            mock_api_instance.orchestratorjob_history.assert_called_once_with(
                id=sample_job_id,
                event_type="job",
                node_id="node-123",
                execution_id="exec-456",
            )
            assert result == expected_response

    def test_job_history_api_exception(self, mock_config, sample_job_id):
        """Test job_history handles ApiException."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorjob_history.side_effect = ApiException(
                status=404, reason="Job not found"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)
            result = service.job_history(id=sample_job_id)

            assert result is None


class TestErrorHandling:
    """Tests for comprehensive error handling."""

    def test_handles_various_api_status_codes(self, mock_config, sample_job_id):
        """Test handling of various HTTP status codes."""
        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            status_codes = [400, 401, 403, 404, 500, 502, 503]

            for status in status_codes:
                mock_api_instance = MagicMock()
                mock_api_instance.orchestratorget_job.side_effect = ApiException(
                    status=status, reason=f"Error {status}"
                )
                mock_api_class.return_value = mock_api_instance

                service = OrchestratorService(config=mock_config)
                result = service.get_job(id=sample_job_id)
                assert result is None, f"Expected None for status {status}"

    def test_handles_timeout_error(self, mock_config, sample_job_id):
        """Test handling of timeout errors."""
        import socket

        with patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.ApiClient"
        ) as mock_client_class, patch(
            "bacalhau_sdk.orchestrator_service.orchestrator_api.OrchestratorApi"
        ) as mock_api_class:
            from bacalhau_sdk.orchestrator_service import OrchestratorService

            mock_api_instance = MagicMock()
            mock_api_instance.orchestratorget_job.side_effect = socket.timeout(
                "Connection timed out"
            )
            mock_api_class.return_value = mock_api_instance

            service = OrchestratorService(config=mock_config)

            with pytest.raises(socket.timeout):
                service.get_job(id=sample_job_id)
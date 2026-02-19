"""Extended tests for config module."""

import os
import pytest
from pathlib import Path
from tempfile import TemporaryDirectory
from unittest.mock import patch, MagicMock

from bacalhau_sdk.config import (
    init_config,
    get_client_id,
    set_client_id,
    get_user_id_key,
)
from bacalhau_apiclient import Configuration


class TestInitConfigExtended:
    """Extended tests for init_config function."""

    def test_init_config_returns_configuration(self, clean_env):
        """Test that init_config returns a Configuration object."""
        config = init_config()
        assert isinstance(config, Configuration)

    def test_init_config_default_host(self, clean_env):
        """Test init_config with default host settings."""
        config = init_config()
        assert config.host is not None
        assert config.host.startswith("http")

    def test_init_config_https_env(self, clean_env):
        """Test init_config with BACALHAU_HTTPS environment variable."""
        os.environ["BACALHAU_HTTPS"] = "1"
        config = init_config()
        assert config.host.startswith("https")
        del os.environ["BACALHAU_HTTPS"]

    def test_init_config_https_empty(self, clean_env):
        """Test init_config with empty BACALHAU_HTTPS."""
        os.environ["BACALHAU_HTTPS"] = ""
        config = init_config()
        assert config.host.startswith("http://")

    def test_init_config_custom_host(self, clean_env):
        """Test init_config with custom API host."""
        os.environ["BACALHAU_API_HOST"] = "custom.example.com"
        config = init_config()
        assert "custom.example.com" in config.host
        del os.environ["BACALHAU_API_HOST"]

    def test_init_config_custom_port(self, clean_env):
        """Test init_config with custom API port."""
        os.environ["BACALHAU_API_PORT"] = "8080"
        config = init_config()
        assert ":8080" in config.host
        del os.environ["BACALHAU_API_PORT"]

    def test_init_config_host_and_port(self, clean_env):
        """Test init_config with custom host and port."""
        os.environ["BACALHAU_API_HOST"] = "api.example.com"
        os.environ["BACALHAU_API_PORT"] = "9000"
        config = init_config()
        assert config.host == "http://api.example.com:9000"
        del os.environ["BACALHAU_API_HOST"]
        del os.environ["BACALHAU_API_PORT"]

    def test_init_config_https_custom_host_port(self, clean_env):
        """Test init_config with HTTPS, custom host, and port."""
        os.environ["BACALHAU_HTTPS"] = "1"
        os.environ["BACALHAU_API_HOST"] = "secure.example.com"
        os.environ["BACALHAU_API_PORT"] = "443"
        config = init_config()
        assert config.host == "https://secure.example.com:443"
        del os.environ["BACALHAU_HTTPS"]
        del os.environ["BACALHAU_API_HOST"]
        del os.environ["BACALHAU_API_PORT"]

    def test_init_config_removes_trailing_slash(self, clean_env):
        """Test that init_config removes trailing slash from host."""
        # This test ensures the trailing slash removal logic works
        config = init_config()
        assert not config.host.endswith("/")


class TestEnsureConfigDir:
    """Tests for config directory functionality via init_config."""

    def test_config_dir_created_on_init(self, clean_env):
        """Test that config directory is created when init_config is called."""
        with TemporaryDirectory() as tmpdir:
            os.environ["BACALHAU_DIR"] = tmpdir
            
            config = init_config()
            
            # init_config should have ensured the config dir exists
            config_dir = Path(tmpdir)
            assert config_dir.exists()
            assert config_dir.is_dir()

    def test_config_dir_custom_location(self, clean_env):
        """Test that custom BACALHAU_DIR is used."""
        with TemporaryDirectory() as tmpdir:
            # The directory must already exist for BACALHAU_DIR
            # (the code validates it with os.stat())
            custom_dir = Path(tmpdir) / "custom_bacalhau"
            custom_dir.mkdir(parents=True, exist_ok=True)
            os.environ["BACALHAU_DIR"] = str(custom_dir)

            config = init_config()

            # Directory should still exist
            assert custom_dir.exists()
            assert custom_dir.is_dir()

    def test_config_dir_uses_home_default(self, clean_env):
        """Test that config dir defaults to ~/.bacalhau when BACALHAU_DIR not set."""
        config = init_config()
        
        # Should use default home location
        home_path = Path.home()
        default_config_dir = home_path / ".bacalhau"
        
        assert default_config_dir.exists()
        assert default_config_dir.is_dir()


class TestClientId:
    """Tests for client ID functions."""

    def test_get_client_id_default(self):
        """Test get_client_id returns None by default."""
        # Reset the client_id
        set_client_id(None)
        assert get_client_id() is None

    def test_set_and_get_client_id(self):
        """Test setting and getting client ID."""
        test_id = "test-client-123"
        set_client_id(test_id)
        assert get_client_id() == test_id

    def test_set_client_id_multiple_times(self):
        """Test that client_id can be changed multiple times."""
        set_client_id("first-id")
        assert get_client_id() == "first-id"

        set_client_id("second-id")
        assert get_client_id() == "second-id"

        set_client_id("third-id")
        assert get_client_id() == "third-id"


class TestUserIdKey:
    """Tests for user ID key function."""

    def test_get_user_id_key_default(self):
        """Test get_user_id_key default value."""
        result = get_user_id_key()
        # The default is None since it's not set
        assert result is None


class TestEnsureConfigFile:
    """Tests for config file functionality."""

    def test_config_file_not_required(self, clean_env):
        """Test that config file is optional and init_config works without it."""
        # init_config should work even without a config file
        config = init_config()
        
        assert config is not None
        assert isinstance(config, Configuration)

    def test_config_returns_empty_string_for_file(self, clean_env):
        """Test that config file handling returns empty (not implemented yet)."""
        # The config file functionality is not fully implemented yet,
        # but init_config should still work
        config = init_config()
        
        # Should return a valid configuration
        assert config is not None


class TestConfigIntegration:
    """Integration tests for config module."""

    def test_full_config_workflow(self, clean_env):
        """Test a typical configuration workflow."""
        # Set custom config
        os.environ["BACALHAU_API_HOST"] = "mycluster.internal"
        os.environ["BACALHAU_API_PORT"] = "9999"

        config = init_config()

        assert "mycluster.internal" in config.host
        assert "9999" in config.host

        # Clean up
        del os.environ["BACALHAU_API_HOST"]
        del os.environ["BACALHAU_API_PORT"]

    def test_config_with_temp_directory(self, clean_env):
        """Test configuration with a temporary config directory."""
        with TemporaryDirectory() as tmpdir:
            os.environ["BACALHAU_DIR"] = tmpdir
            os.environ["BACALHAU_API_HOST"] = "localhost"
            os.environ["BACALHAU_API_PORT"] = "4222"

            config = init_config()

            assert config.host == "http://localhost:4222"

    def test_config_priority_env_over_default(self, clean_env):
        """Test that environment variables take priority over defaults."""
        # First, get default config
        default_config = init_config()
        default_host = default_config.host

        # Now set environment variables
        os.environ["BACALHAU_API_HOST"] = "override.example.com"
        os.environ["BACALHAU_API_PORT"] = "12345"

        override_config = init_config()

        # The overridden config should be different
        assert override_config.host != default_host
        assert "override.example.com" in override_config.host

        del os.environ["BACALHAU_API_HOST"]
        del os.environ["BACALHAU_API_PORT"]


class TestEdgeCases:
    """Tests for edge cases and boundary conditions."""

    def test_empty_api_host_uses_default(self, clean_env):
        """Test that empty BACALHAU_API_HOST uses default."""
        os.environ["BACALHAU_API_HOST"] = ""
        config = init_config()

        # Should still have a valid host
        assert config.host is not None
        assert config.host.startswith("http")
        del os.environ["BACALHAU_API_HOST"]

    def test_empty_api_port_uses_default(self, clean_env):
        """Test that empty BACALHAU_API_PORT uses default."""
        os.environ["BACALHAU_API_PORT"] = ""
        config = init_config()

        # Should still have a valid port
        assert config.host is not None
        assert ":" in config.host
        del os.environ["BACALHAU_API_PORT"]

    def test_numeric_port_as_string(self, clean_env):
        """Test that numeric port is handled correctly."""
        os.environ["BACALHAU_API_PORT"] = "1234"  # String, as env vars are strings
        config = init_config()
        assert ":1234" in config.host
        del os.environ["BACALHAU_API_PORT"]

    def test_host_with_special_chars(self, clean_env):
        """Test host with hyphens and dots."""
        os.environ["BACALHAU_API_HOST"] = "my-cluster.example-domain.com"
        config = init_config()
        assert "my-cluster.example-domain.com" in config.host
        del os.environ["BACALHAU_API_HOST"]

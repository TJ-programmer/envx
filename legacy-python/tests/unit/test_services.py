from __future__ import annotations

import pytest

from envx.cli.commands.common import build_service
from envx.core.exceptions import EnvironmentConflictError, EnvironmentNotFoundError
from envx.security.redaction import REDACTED_VALUE


def test_secret_values_are_redacted_by_default(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    service = build_service()
    service.init_project()
    service.set_variable("API_KEY", "super-secret", secret=True)

    rows = service.list_variables()

    assert rows[0]["value"] == REDACTED_VALUE


def test_show_secrets_reveals_plaintext(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    service = build_service()
    service.init_project()
    service.set_variable("API_KEY", "super-secret", secret=True)

    rows = service.list_variables(show_secrets=True)

    assert rows[0]["value"] == "super-secret"


def test_environment_lifecycle(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    service = build_service()
    service.init_project()
    service.create_environment("prod")
    service.use_environment("prod")

    rows = service.list_environments()

    assert any(row["name"] == "prod" and row["active"] == "true" for row in rows)


def test_delete_active_environment_is_rejected(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    service = build_service()
    service.init_project()

    with pytest.raises(EnvironmentConflictError):
        service.delete_environment("dev")


def test_missing_environment_raises(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    service = build_service()
    service.init_project()

    with pytest.raises(EnvironmentNotFoundError):
        service.use_environment("prod")

from __future__ import annotations

import json
import subprocess
import sys

from typer.testing import CliRunner

from envx.cli.app import app

runner = CliRunner()


def test_init_and_set_secret(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)

    result = runner.invoke(app, ["init"])
    assert result.exit_code == 0

    result = runner.invoke(app, ["set", "API_KEY", "my-secret", "--secret"])
    assert result.exit_code == 0

    config = json.loads((tmp_path / ".envx" / "config.json").read_text(encoding="utf-8"))
    stored = config["environments"]["dev"]["variables"]["API_KEY"]["value"]
    assert stored.startswith("enc:")


def test_list_redacts_by_default_and_can_show_secrets(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    runner.invoke(app, ["init"])
    runner.invoke(app, ["set", "API_KEY", "my-secret", "--secret"])

    listed = runner.invoke(app, ["list"])
    shown = runner.invoke(app, ["list", "--show-secrets"])

    assert "********" in listed.stdout
    assert "my-secret" not in listed.stdout
    assert "my-secret" in shown.stdout


def test_run_injects_environment_vector_mode(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    runner.invoke(app, ["init"])
    runner.invoke(app, ["set", "PORT", "8123"])
    output_path = tmp_path / "vector-output.txt"

    result = runner.invoke(
        app,
        [
            "run",
            "--",
            sys.executable,
            "-c",
            f"from pathlib import Path; import os; Path(r'{output_path}').write_text(os.environ['PORT'], encoding='utf-8')",
        ],
    )

    assert result.exit_code == 0
    assert output_path.read_text(encoding="utf-8") == "8123"


def test_run_injects_environment_shell_mode(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    runner.invoke(app, ["init"])
    runner.invoke(app, ["set", "SHELL_SECRET", "works"])
    output_path = tmp_path / "shell-output.txt"
    shell_command = subprocess.list2cmdline(
        [
            sys.executable,
            "-c",
            f"from pathlib import Path; import os; Path(r'{output_path}').write_text(os.environ['SHELL_SECRET'], encoding='utf-8')",
        ]
    )

    result = runner.invoke(app, ["run", "--shell", shell_command])

    assert result.exit_code == 0
    assert output_path.read_text(encoding="utf-8") == "works"


def test_corrupt_config_returns_clear_error(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    envx_dir = tmp_path / ".envx"
    envx_dir.mkdir()
    (envx_dir / "config.json").write_text("{", encoding="utf-8")

    result = runner.invoke(app, ["list"])

    assert result.exit_code == 1
    assert "invalid JSON" in result.stderr


def test_missing_key_fails_safely(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    runner.invoke(app, ["init"])
    runner.invoke(app, ["set", "API_KEY", "my-secret", "--secret"])
    (tmp_path / ".envx" / "key.bin").unlink()

    result = runner.invoke(app, ["list", "--show-secrets"])

    assert result.exit_code == 1
    assert "Encryption key not found" in result.stderr


def test_environment_commands_work(tmp_path, monkeypatch):
    monkeypatch.chdir(tmp_path)
    runner.invoke(app, ["init"])

    created = runner.invoke(app, ["env", "create", "staging"])
    switched = runner.invoke(app, ["env", "use", "staging"])
    listed = runner.invoke(app, ["env", "list"])

    assert created.exit_code == 0
    assert switched.exit_code == 0
    assert "staging" in listed.stdout
    assert "true" in listed.stdout

from __future__ import annotations

import json

import pytest

from envx.config.paths import resolve_project_paths
from envx.core.exceptions import ConfigCorruptedError, SecretKeyError
from envx.core.models import ConfigFile
from envx.security.crypto import CryptoManager
from envx.security.keyring import LocalKeyProvider
from envx.storage.json_store import JsonConfigStore


def build_store(tmp_path):
    paths = resolve_project_paths(tmp_path)
    store = JsonConfigStore(paths)
    crypto = CryptoManager(LocalKeyProvider(paths))
    return paths, store, crypto


def test_config_round_trip(tmp_path):
    paths, store, _ = build_store(tmp_path)
    config = ConfigFile.create_default()
    store.save(config)

    loaded = store.load()

    assert loaded.version == 1
    assert loaded.active_env == "dev"
    assert "dev" in loaded.environments
    assert paths.config_path.exists()


def test_backup_is_written_on_second_save(tmp_path):
    paths, store, _ = build_store(tmp_path)
    config = ConfigFile.create_default()
    store.save(config)
    config.metadata["updated"] = True
    store.save(config)

    assert paths.backup_path.exists()
    backup = json.loads(paths.backup_path.read_text(encoding="utf-8"))
    assert backup["active_env"] == "dev"


def test_legacy_config_is_migrated(tmp_path):
    paths, store, _ = build_store(tmp_path)
    paths.config_dir.mkdir(parents=True, exist_ok=True)
    paths.config_path.write_text(
        json.dumps({"active_env": "dev", "environments": {"dev": {"PORT": "8000"}}}),
        encoding="utf-8",
    )

    config = store.load()

    assert config.version == 1
    assert config.environments["dev"].variables["PORT"].value == "8000"


def test_corrupt_config_raises_actionable_error(tmp_path):
    paths, store, _ = build_store(tmp_path)
    paths.config_dir.mkdir(parents=True, exist_ok=True)
    paths.config_path.write_text("{", encoding="utf-8")

    with pytest.raises(ConfigCorruptedError):
        store.load()


def test_crypto_invalid_key_fails_closed(tmp_path):
    paths, _, crypto = build_store(tmp_path)
    paths.config_dir.mkdir(parents=True, exist_ok=True)
    paths.key_path.write_bytes(b"broken")

    with pytest.raises(SecretKeyError):
        crypto.decrypt("enc:abc")

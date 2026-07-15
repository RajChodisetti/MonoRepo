"""Tests for ingestion environment precedence."""

import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import env_loader


class EnvLoaderTest(unittest.TestCase):
    def test_explicit_ingestion_file_wins_but_process_environment_is_preserved(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            backend = root / "backend"
            outreach = root / "automation" / "outreach"
            backend.mkdir()
            outreach.mkdir(parents=True)
            (root / ".env").write_text("VALUE=root\nPROCESS_VALUE=root\n", encoding="utf-8")
            (backend / ".env").write_text("VALUE=backend\n", encoding="utf-8")
            (outreach / ".env").write_text("VALUE=outreach\n", encoding="utf-8")
            explicit = root / "ingestion.env"
            explicit.write_text("VALUE=explicit\n", encoding="utf-8")

            with (
                patch.object(env_loader, "_MONOREPO", root),
                patch.object(env_loader, "_OUTREACH", outreach),
                patch.dict(
                    os.environ,
                    {
                        "INGESTION_ENV_FILE": str(explicit),
                        "PROCESS_VALUE": "process",
                    },
                    clear=True,
                ),
            ):
                env_loader.load_project_env()
                self.assertEqual(os.environ["VALUE"], "explicit")
                self.assertEqual(os.environ["PROCESS_VALUE"], "process")


if __name__ == "__main__":
    unittest.main()

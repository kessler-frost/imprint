import logging
import pytest
import subprocess
from pathlib import Path

logger = logging.getLogger(__name__)


@pytest.fixture(scope="session")
def project_root():
    """Return the project root directory."""
    return Path(__file__).parent.parent


@pytest.fixture(scope="session")
def imprint_binary(project_root):
    """Build imprint binary once for all tests.

    Returns: Absolute path to the imprint binary as a string.
    """
    bin_path = project_root / "bin" / "imprint"
    result = subprocess.run(
        ["go", "build", "-o", str(bin_path), "./cmd/imprint"],
        cwd=project_root,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        pytest.fail(
            f"Failed to build imprint binary.\n"
            f"stdout: {result.stdout}\n"
            f"stderr: {result.stderr}"
        )
    return str(bin_path.resolve())


@pytest.fixture(scope="session")
def example_binaries(project_root):
    """Build all example binaries.

    Returns: Dict mapping example name to absolute binary path.
    """
    examples_dir = project_root / "examples"
    if not examples_dir.exists():
        pytest.fail(f"Examples directory not found: {examples_dir}")

    binaries = {}
    expected_examples = ["text-demo", "screenshot-demo", "what-changed"]

    for example in expected_examples:
        example_path = examples_dir / example
        if not example_path.exists():
            pytest.fail(f"Example directory not found: {example_path}")

        result = subprocess.run(
            ["go", "build", "-o", example, "."],
            cwd=example_path,
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            pytest.fail(
                f"Failed to build {example}.\n"
                f"stdout: {result.stdout}\n"
                f"stderr: {result.stderr}"
            )
        binaries[example] = str((example_path / example).resolve())
    return binaries

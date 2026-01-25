import pytest
import subprocess
from pathlib import Path


@pytest.fixture(scope="session")
def project_root():
    """Return the project root directory."""
    return Path(__file__).parent.parent


@pytest.fixture(scope="session")
def imprint_binary(project_root):
    """Build imprint binary once for all tests."""
    bin_path = project_root / "bin" / "imprint"
    subprocess.run(
        ["go", "build", "-o", str(bin_path), "./cmd/imprint"],
        cwd=project_root,
        check=True,
    )
    return str(bin_path.resolve())


@pytest.fixture(scope="session")
def example_binaries(project_root):
    """Build all example binaries."""
    examples_dir = project_root / "examples"
    binaries = {}
    for example in ["text-demo", "screenshot-demo", "what-changed"]:
        example_path = examples_dir / example
        subprocess.run(
            ["go", "build", "-o", example, "."],
            cwd=example_path,
            check=True,
        )
        binaries[example] = str((example_path / example).resolve())
    return binaries

import pytest

def pytest_addoption(parser):
    parser.addoption(
        "--bin-path", action="store", default=None, help="Path to the binary file."
    )

@pytest.fixture
def bin_path(request):
    return request.config.getoption("--bin-path")
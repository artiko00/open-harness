from app.health import health


def test_health():
    assert health() == "ok"

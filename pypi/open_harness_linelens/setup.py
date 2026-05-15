"""Force a platform-specific wheel so pip picks the right native binary."""
from setuptools import setup
from setuptools.dist import Distribution


class BinaryDistribution(Distribution):
    """Marks the package as containing platform-specific binaries.

    This makes setuptools emit a wheel with a real platform tag (e.g.
    manylinux_2_17_x86_64) instead of the universal `py3-none-any` tag.
    The actual platform tag is set via `--plat-name` from build-pypi.sh.
    """

    def has_ext_modules(self):
        return True


setup(distclass=BinaryDistribution)

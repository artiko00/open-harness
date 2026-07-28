"""Build a platform-specific but Python-version-agnostic wheel.

The wheel embeds a native Go binary, so it must be platform-tagged (e.g.
manylinux2014_x86_64) but the binary does not link against any Python ABI.
We override bdist_wheel to emit `py3-none-<plat>` instead of the default
`cp3X-cp3X-<plat>` that BinaryDistribution would otherwise force.
"""
from setuptools import setup
from setuptools.dist import Distribution

try:
    from wheel.bdist_wheel import bdist_wheel as _bdist_wheel
except ImportError:
    _bdist_wheel = None


class BinaryDistribution(Distribution):
    def has_ext_modules(self):
        return True


cmdclass = {}
if _bdist_wheel is not None:

    class bdist_wheel_any_python(_bdist_wheel):
        def finalize_options(self):
            super().finalize_options()
            self.root_is_pure = False  # platform-specific wheel

        def get_tag(self):
            _, _, plat = super().get_tag()
            return ("py3", "none", plat)

    cmdclass["bdist_wheel"] = bdist_wheel_any_python


setup(distclass=BinaryDistribution, cmdclass=cmdclass)

from setuptools import setup


setup(
    name="bc-data-analyzer",
    packages=["bc_data_analyzer"],
    scripts=["bin/bc-data-analyzer"],
    include_package_data=True,
    install_requires=[
        "numpy",
        "matplotlib",
    ],
    extras_require={
        "dev": [
            "pylint",
            "ipython",
        ]
    }
)

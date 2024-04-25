import sys

from invoke import task


@task
def invoke_unit_tests(ctx):
    """
    Run the unit tests on the invoke tasks
    """
    ctx.run(
        f"mutatest -t '{sys.executable} -m unittest discover -s tasks -p \"*_tests.py\"'",
        env={"GITLAB_TOKEN": "fake_token"},
    )

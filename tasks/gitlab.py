import os
import re

import yaml
from invoke import task

from tasks.libs.ciproviders.gitlab import read_content

DO_NOT_RUN = [{'when': 'never'}]
RUN_ON_SUCCESS = [{'when': 'on_success'}]
# RUN_MANUAL = [{'when': 'manual', 'allow_failure': 'true'}]
# RUN_ALWAYS = [{'when': 'always'}]


@task
def generate_pipeline(ctx, base="main", output_file="generated.yml"):
    contents = read_content(".gitlab-ci-base.yml")

    contents[".on_packaging_change"] = _on_packaging_change(ctx, base)

    with open(output_file, 'w+') as f:
        yaml.safe_dump(contents, f)


# Decisions


def _on_packaging_change(ctx, base):
    if _is_mergequeue(ctx):
        return DO_NOT_RUN

    if _is_modified(
        ctx,
        base,
        [
            r"^omnibus/",
            r"^\.gitlab-ci\.yml$",
            r"^\.gitlab/package_build\.yml$",
            r"^\.gitlab/package_build/",
            r"^release\.json$",
        ],
    ):
        return RUN_ON_SUCCESS

    return DO_NOT_RUN


def _is_mergequeue(ctx):
    if re.match("^mq-working-branch-", _get_git_branch(ctx)):
        return True
    return False


def _is_modified(ctx, base, rules):
    pattern = "|".join(f"({rule})" for rule in rules)
    print(pattern)
    regex = re.compile(pattern)
    for file in _get_modified_files(ctx, base):
        if regex.match(file):
            print(f"Matched {file}")
            return True
    return False


# Functions to get data


def _get_modified_files(ctx, base):
    last_main_commit = ctx.run(f"git merge-base HEAD origin/{base}", hide=True).stdout

    modified_files = ctx.run(f"git diff --name-only --no-renames {last_main_commit}", hide=True).stdout.splitlines()
    return modified_files


def _get_git_branch(ctx):
    if os.environ.get("OVERRIDE_GIT_BRANCH"):
        return os.environ.get("OVERRIDE_GIT_BRANCH")

    if os.environ.get("CI_COMMIT_BRANCH"):
        return os.environ.get("CI_COMMIT_BRANCH")

    return ctx.run("git rev-parse --abbrev-ref HEAD", hide=True).stdout

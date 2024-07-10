from __future__ import annotations

import io
import os
import traceback
from datetime import timedelta

from invoke import Context, task
from invoke.exceptions import Exit

import tasks.libs.notify.unit_tests as unit_tests_utils
from tasks.libs.ciproviders.gitlab_api import (
    get_gitlab_ci_configuration,
    print_gitlab_ci_configuration,
)
from tasks.libs.common.datadog_api import send_metrics
from tasks.libs.notify import alerts, failure_summary, pipeline_status
from tasks.libs.notify.utils import PROJECT_NAME
from tasks.libs.pipeline.data import get_failed_jobs
from tasks.libs.pipeline.notifications import (
    base_message,
    check_for_missing_owners_slack_and_jira,
)
from tasks.libs.pipeline.stats import compute_failed_jobs_series, compute_required_jobs_max_duration


@task
def check_teams(_):
    if check_for_missing_owners_slack_and_jira():
        print(
            "Error: Some teams in CODEOWNERS don't have their slack notification channel or jira specified!\n"
            "Please specify one in the GITHUB_SLACK_MAP or GITHUB_JIRA_MAP maps in tasks/libs/pipeline/github_slack_map.yaml"
            " or tasks/libs/pipeline/github_jira_map.yaml"
        )
        raise Exit(code=1)
    else:
        print("All CODEOWNERS teams have their slack notification channel and jira project specified !!")


@task
def send_message(ctx: Context, notification_type: str = "merge", dry_run: bool = False):
    """
    Send notifications for the current pipeline. CI-only task.
    Use the --dry-run option to test this locally, without sending
    real slack messages.
    """

    try:
        failed_jobs = get_failed_jobs(PROJECT_NAME, os.environ["CI_PIPELINE_ID"])
        messages_to_send = pipeline_status.generate_failure_messages(PROJECT_NAME, failed_jobs)
    except Exception as e:
        buffer = io.StringIO()
        print(base_message("datadog-agent", "is in an unknown state"), file=buffer)
        print("Found exception when generating notification:", file=buffer)
        traceback.print_exc(limit=-1, file=buffer)
        print("See the notify job log for the full exception traceback.", file=buffer)

        # Print traceback on job log
        print(e)
        traceback.print_exc()
        raise Exit(code=1) from e

    pipeline_status.send_message_and_metrics(ctx, failed_jobs, messages_to_send, notification_type, dry_run)


@task
def send_stats(_, dry_run=False):
    """
    Send statistics to Datadog for the current pipeline. CI-only task.
    Use the --dry-run option to test this locally, without sending
    data points to Datadog.
    """
    if not (dry_run or os.environ.get("DD_API_KEY")):
        print("DD_API_KEY environment variable not set, cannot send pipeline metrics to the backend")
        raise Exit(code=1)

    series = compute_failed_jobs_series(PROJECT_NAME)
    series.extend(compute_required_jobs_max_duration(PROJECT_NAME))

    if not dry_run:
        send_metrics(series)
        print(f"Sent pipeline metrics: {series}")
    else:
        print(f"Would send: {series}")


@task
def check_consistent_failures(ctx, job_failures_file="job_executions.v2.json"):
    # Retrieve the stored document in aws s3. It has the following format:
    # {
    #     "pipeline_id": 123,
    #     "jobs": {
    #         "job1": {"consecutive_failures": 2, "jobs_info": [{"id": null, "failing": false, "commit": "abcdef42"}, {"id": 314618, "failing": true, "commit": "abcdef42"}, {"id": 618314, "failing": true, "commit": "abcdef42"}]},
    #         "job2": {"consecutive_failures": 0, "cumulative_failures": [{"id": 314618, "failing": true, "commit": "abcdef42"}, {"id": null, "failing": false, "commit": "abcdef42"}]},
    #         "job3": {"consecutive_failures": 1, "cumulative_failures": [{"id": 314618, "failing": true, "commit": "abcdef42"}]},
    #     }
    # }
    # NOTE: this format is described by the Executions class
    # The pipeline_id is used to by-pass the check if the pipeline chronological order is not respected
    # The jobs dictionary contains the consecutive and cumulative failures for each job
    # The consecutive failures are reset to 0 when the job is not failing, and are raising an alert when reaching the CONSECUTIVE_THRESHOLD (3)
    # The cumulative failures list contains 1 for failures, 0 for succes. They contain only then CUMULATIVE_LENGTH(10) last executions and raise alert when 50% failure rate is reached

    job_executions = alerts.retrieve_job_executions(ctx, job_failures_file)

    # By-pass if the pipeline chronological order is not respected
    if job_executions.pipeline_id > int(os.environ["CI_PIPELINE_ID"]):
        return
    job_executions.pipeline_id = int(os.environ["CI_PIPELINE_ID"])

    alert_jobs, job_executions = alerts.update_statistics(job_executions)

    alerts.send_notification(ctx, alert_jobs)

    alerts.upload_job_executions(ctx, job_executions, job_failures_file)


@task
def failure_summary_upload_pipeline_data(ctx):
    """
    Upload failure summary data to S3 at the end of each main pipeline
    """
    failure_summary.upload_summary(ctx, os.environ['CI_PIPELINE_ID'])


@task
def failure_summary_send_notifications(
    ctx, daily_summary: bool = False, weekly_summary: bool = False, max_length: int = 8
):
    """
    Make summaries from data in s3 and send them to slack
    """

    assert (
        daily_summary or weekly_summary and not (daily_summary and weekly_summary)
    ), "Only one of daily or weekly summary can be set"

    period = timedelta(days=1) if daily_summary else timedelta(weeks=1)
    failure_summary.send_summary_messages(ctx, weekly_summary, max_length, period)


@task
def unit_tests(ctx, pipeline_id, pipeline_url, branch_name, dry_run=False):
    jobs_with_no_tests_run = unit_tests_utils.process_unit_tests_tarballs(ctx)
    msg = unit_tests_utils.create_msg(pipeline_id, pipeline_url, jobs_with_no_tests_run)

    if dry_run:
        print(msg)
    else:
        unit_tests_utils.comment_pr(msg, pipeline_id, branch_name, jobs_with_no_tests_run)


@task
def print_gitlab_ci(
    ctx,
    input_file: str = '.gitlab-ci.yml',
    job: str | None = None,
    sort: bool = False,
    clean: bool = True,
    git_ref: str | None = None,
    ignore_errors: bool = False,
):
    """
    Prints the full gitlab ci configuration.

    - job: If provided, print only one job
    - clean: Apply post processing to make output more readable (remove extends, flatten lists of lists...)
    - ignore_errors: If True, ignore errors in the gitlab configuration (only process yaml)
    - git_ref: If provided, use this git reference to fetch the configuration
    """

    yml = get_gitlab_ci_configuration(
        ctx, input_file, job=job, clean=clean, git_ref=git_ref, ignore_errors=ignore_errors
    )

    # Print
    print_gitlab_ci_configuration(yml, sort_jobs=sort)

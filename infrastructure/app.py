"""Define the container registry used by flows2fim releases."""

import aws_cdk as cdk
from aws_cdk import aws_ecr as ecr

app = cdk.App()
stack = cdk.Stack(
    app,
    "Flows2fimRegistry",
    env=cdk.Environment(
        account=cdk.Aws.ACCOUNT_ID,
        region=cdk.Aws.REGION,
    ),
    synthesizer=cdk.DefaultStackSynthesizer(generate_bootstrap_version_rule=False),
)

repository = ecr.Repository(
    stack,
    "Repository",
    repository_name="flows2fim",
    image_scan_on_push=True,
    image_tag_mutability=ecr.TagMutability.IMMUTABLE,
    encryption=ecr.RepositoryEncryption.AES_256,
    removal_policy=cdk.RemovalPolicy.RETAIN,
    lifecycle_rules=[
        ecr.LifecycleRule(
            description="Keep the 100 newest commit images",
            tag_prefix_list=["sha-"],
            max_image_count=100,
        ),
        ecr.LifecycleRule(
            description="Remove untagged images after seven days",
            tag_status=ecr.TagStatus.UNTAGGED,
            max_image_age=cdk.Duration.days(7),
        ),
    ],
)

cdk.CfnOutput(stack, "RepositoryName", value=repository.repository_name)
cdk.CfnOutput(stack, "RepositoryUri", value=repository.repository_uri)

app.synth()

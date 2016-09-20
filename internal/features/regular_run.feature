@regular
Feature: Normal Run
  As a normal user
  I want to perform a backup to a remote S3 system

  Scenario: Successfully backs up single directory
    Given this is not a dry run
    When I run the backup for a single directory
    Then I should see the expected files on the s3 host
    And I should see the following output:
    """
      Test
    """

  Scenario: Successfully backs up multiple directories

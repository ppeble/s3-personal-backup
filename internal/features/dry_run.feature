Feature: Dry Run
  As a normal user
  I want to perform a dry run of a backup to a remote S3 system

  Scenario: Successfully prints output of single directory
    Given this is a dry run
    When I run the backup for a single directory
    Then I should see the expected files on the s3 host
    And I should see the following output:
    """
      Test
    """

  Scenario: Successfully prints output of multiple directories

# dreamhost-personal-backup [![Build Status](https://travis-ci.org/ptrimble/dreamhost-personal-backup.svg?branch=master)](https://travis-ci.org/ptrimble/dreamhost-personal-backup)

A personal backup tool written in Go. This project writes files to your favorite S3-compatible
storage service.

This project serves as an excuse for me to solve a problem in Go. I know that
I could use any number of existing products to address my backup needs but this is
way more fun.

*Please* do not use this project for anything that is mission-critical. I back up
my music and personal documents to a remote server as another duplicate in a myriad
of backup locations. I don't rely on just this backup.

There is no versioning. This does not back up symlinks or directories. It simply
walks recursively through from the supplied directory and pushes up every file to
the remote S3 storage. That's it!

## Usage

<need to add `go get` stuff here, not ready for it yet>

### Configuration

This project assumes that you have already set up a user on your preferred S3-compatible host.

The `dreamhost_personal_backup` binary requires the following information:

* root directory - specified via the `-targetDir <dir>` flag or by setting the `PERSONAL_BACKUP_TARGET_DIR` env variable
* S3 host - specified via the `s3Host <host>` flag or by setting the `PERSONAL_BACKUP_S3_HOST` env variable
* S3 access key - specified via the `s3AccessKey <key>` flag or by setting the `PERSONAL_BACKUP_S3_ACCESS_KEY` env variable
* S3 secret key - specified via the `s3SecretKey <key>` flag or by setting the `PERSONAL_BACKUP_S3_SECRET_KEY` env variable
* S3 bucket name - specified via the `-s3BucketName` flag or the `PERSONAL_BACKUP_S3_BUCKET_NAME` env variable

In addition, there are optional fields:

* remote worker count - number of workers to run in parallel to process actions on the remote host. Used currently to (primitively) limit bandwidth usage. Fewer workers means fewer simultaneous actions (like uploading) run against the S3 host. Specified via the `-remoteWorkerCount` flag or the `PERSONAL_BACKUP_REMOTE_WORKER_COUNT` env variable

In all instances the command line flag will take priority over the environment variable.

## TODO

* Ability to print report of specific directories/files and their status on the remote host. Are they backed up?
* Ability to pull down files either selectively or as a whole from remote S3 instance to local
* Progress reporting on transfers
* Ability to read from config file as well as command line and env vars
* Ability to leave S3 Bucket blank and have one be generated. It should then be saved in the config so it is consistently referenced.

## Credits

* I used the [minio-go](https://github.com/minio/minio-go) client
* I used the documentation found at the [DreamObjects Wiki](http://wiki.dreamhost.com/DreamObjects_Overview_and_FAQs) for system-specific information on Dreamhost infrastructure

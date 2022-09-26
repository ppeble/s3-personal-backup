# s3-personal-backup

A personal backup tool written in Go. This project writes files to your favorite S3-compatible
storage service.

This project serves as an excuse for me to solve a problem in Go. I know that
I could use any number of existing products to address my backup needs but this is
way more fun.

What I want is a simple and idempotent "put copies from a directory in s3 recursively"
solution. I don't want encryption. I don't want versioning. I want to point at an s3-compatible
solution and make a copy from my local system for backup purposes so I could download individual
files or directories as it suited me.

There is no versioning. This does not back up symlinks or directories. It simply
walks recursively through from the supplied directory and pushes up every file to
the remote S3 storage. That's it!

*Please* do not use this project for anything that is mission-critical. I back up
my music and personal documents to a remote server as another duplicate in a myriad
of backup locations. I don't rely on just this backup.

## Usage

<need to add `go get` stuff here, not ready for it yet>

### Configuration

This project assumes that you have already set up a user on your preferred S3-compatible host.

The `s3_personal_backup` binary requires the following information:

* backup target directories - specified via the `--targetDirs <dir>` flag or by setting the `PERSONAL_BACKUP_TARGETDIRS` env variable. Should be a comma separated list of full directory paths to back up. Ex: '/home/<user>/documents,/home/<user>/music,/home/<user>/pictures,/media/dir,/etc/dir'
* S3 host - specified via the `--s3Host <host>` flag or by setting the `PERSONAL_BACKUP_S3HOST` env variable
* S3 access key - specified via the `--s3AccessKey <key>` flag or by setting the `PERSONAL_BACKUP_S3ACCESSKEY` env variable
* S3 secret key - specified via the `--s3SecretKey <key>` flag or by setting the `PERSONAL_BACKUP_S3SECRETKEY` env variable
* S3 bucket name - specified via the `--s3BucketName <name>` flag or the `PERSONAL_BACKUP_S3BUCKETNAME` env variable

In addition, there are optional fields:

* remote worker count - DEFAULT 5 - number of workers to run in parallel to process actions on the remote host. Used currently to (primitively) limit bandwidth usage. Fewer workers means fewer simultaneous actions (like uploading) run against the S3 host. Specified via the `--remoteWorkerCount <count>` flag or the `PERSONAL_BACKUP_REMOTEWORKERCOUNT` env variable

In all instances the command line flag will take priority over the environment variable.

## TODO

* Ability to print report of specific directories/files and their status on the remote host. Are they backed up?
* Ability to pull down files either selectively or as a whole from remote S3 instance to local
* Progress reporting on transfers

## Credits

* I used the [minio-go](https://github.com/minio/minio-go) client

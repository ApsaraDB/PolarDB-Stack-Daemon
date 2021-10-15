#!/bin/sh


#package version
#
#const GitBranch = "PolarDBCCMGitBranch"
#const GitCommitId = "PolarDBCCMGitCommitId"
#const GitCommitDate = "PolarDBCCMGitCommitDate"


commit_id=$(git rev-parse HEAD)
echo "commitId: $commit_id"

commit_branch=$(git symbolic-ref --short -q HEAD)
echo "branch $commit_branch"

commit_date=$(git log -1 --format="%cd")
echo "git commit date: $commit_date"


gitrepo=$(git remote -v|grep origin|grep fetch|awk '{print $2}')
echo "git repo: $gitrepo"

rm -f version/version.go
echo "package version" > version/version.go
echo "" >>version/version.go
echo "const GitBranch = \"$commit_branch\"" >> version/version.go
echo "const GitCommitId = \"$commit_id\"" >> version/version.go
echo "const GitCommitDate = \"$commit_date\"" >> version/version.go
echo "const GitCommitRepo = \"$gitrepo\"" >> version/version.go
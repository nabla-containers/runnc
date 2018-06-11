bats-core is a git subtree of `bats-core/bats-core:/libexec`. The commands used to create this subtree were (getting the latest changes from bats-core should be done using the same commands):

```
git checkout bats-core/master
git subtree split -P libexec -b temporary-bats-branch
git checkout tests-in-bats
git subtree add --squash -P tests/bats-core temporary-bats-branch
```

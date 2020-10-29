TAG=2.1
COMMIT_MSG=""

git tag -a scrape-${TAG} -m "${COMMIT_MSG}"
git tag -a process-url-${TAG} -m "${COMMIT_MSG}"
git tag -a processor-${TAG} -m "${COMMIT_MSG}"
git tag -a storage-${TAG} -m "${COMMIT_MSG}"
git tag -a builder-${TAG} -m "${COMMIT_MSG}"
git tag -a bot-${TAG} -m "${COMMIT_MSG}"

git push origin --tags

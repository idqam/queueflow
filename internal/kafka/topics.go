package kafka

const (
	TopicJobsHigh   = "jobs.high"
	TopicJobsNormal = "jobs.normal"
	TopicJobsLow    = "jobs.low"
	TopicJobsDLQ    = "jobs.dlq"
)

var PriorityTopicMap = map[string]string{
	"high":   TopicJobsHigh,
	"normal": TopicJobsNormal,
	"low":    TopicJobsLow,
}

func TopicForPriority(priority string) string {
	if t, ok := PriorityTopicMap[priority]; ok {
		return t
	}
	return TopicJobsNormal
}

var AllJobTopics = []string{TopicJobsHigh, TopicJobsNormal, TopicJobsLow}

package pulsar

// ======================================================================
// PREFIX
// ======================================================================

const (
	TopicPrefix = "plans"
)

// ======================================================================
// PLAN
// ======================================================================

const (
	TopicPlanCreated = TopicPrefix + ".plan.created"
	TopicPlanUpdated = TopicPrefix + ".plan.updated"
	TopicPlanDeleted = TopicPrefix + ".plan.deleted"
)
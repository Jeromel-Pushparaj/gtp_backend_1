package migrations

// GetAllMigrations returns all registered migrations
func GetAllMigrations() []Migration {
	return []Migration{
		Migration001Learnings(),
		Migration002FailurePatterns(),
		Migration003Strategies(),
		Migration004ChatSessions(),
		Migration005TestRuns(),
	}
}


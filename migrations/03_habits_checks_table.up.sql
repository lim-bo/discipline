-- +goose: up
CREATE TABLE IF NOT EXISTS habit_checks (
    id SERIAL PRIMARY KEY,
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    check_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(habit_id, check_date)
);

CREATE INDEX idx_habit_checks_habit_id_date ON habit_checks(habit_id, check_date);
CREATE INDEX idx_habit_checks_date ON habit_checks(check_date);
CREATE INDEX idx_habit_checks_habit_id ON habit_checks(habit_id);
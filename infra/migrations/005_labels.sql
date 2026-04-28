-- +goose Up
CREATE TABLE labels (
    id           bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    decision_id  uuid        NOT NULL UNIQUE
                             REFERENCES decisions(id)
                             ON DELETE RESTRICT,
    label        text        NOT NULL
                             CHECK (label IN ('true_positive', 'false_positive', 'false_negative')),
    labeled_at   timestamptz NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE labels IS
    'Analyst verdicts on past decisions. One label per decision.';

COMMENT ON COLUMN labels.label IS
    'true_positive: flagged and malicious. false_positive: flagged but benign. false_negative: missed but malicious.';

-- +goose Down
DROP TABLE IF EXISTS labels;
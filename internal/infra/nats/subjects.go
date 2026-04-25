package nats

// Stream names — must match the Python RAG engine exactly.
const (
	StreamIngest     = "RAG_INGEST"
	StreamIngestDone = "RAG_INGEST_DONE"
	StreamQuiz       = "RAG_QUIZ"
	StreamQuizDone   = "RAG_QUIZ_DONE"
)

// Subject strings — must mirror rag/app/messaging/subjects.py.
const (
	SubjectIngestRequest = "rag.ingest.request"
	SubjectIngestDone    = "rag.ingest.done"
	SubjectQuizRequest   = "quiz.generate.request"
	SubjectQuizDone      = "quiz.generate.done"
)

// Durable consumer names used by Go workers.
const (
	DurableIngestDone = "go-ingest-done"
	DurableQuizDone   = "go-quiz-done"
)

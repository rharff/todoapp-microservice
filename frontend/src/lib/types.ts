export type Task = {
	id: string;
	title: string;
	stage: 'todo' | 'in_progress' | 'review' | 'done';
	position: number;
	created_at: string;
};

export type AuditLog = {
	id: number;
	task_id: string | null;
	action_string: string;
	created_at: string;
};

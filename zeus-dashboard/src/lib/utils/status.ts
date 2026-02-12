// ステータス表示の共通ユーティリティ

// Vision ステータス
const VISION_STATUS_MAP: Record<string, { color: string; label: string }> = {
	active: { color: '#22c55e', label: 'Active' },
	draft: { color: '#6b7280', label: 'Draft' },
	archived: { color: '#3b82f6', label: 'Archived' }
};

export function getVisionStatusColor(status: string): string {
	return VISION_STATUS_MAP[status]?.color ?? '#6b7280';
}

export function getVisionStatusLabel(status: string): string {
	return VISION_STATUS_MAP[status]?.label ?? status;
}

// Objective ステータス
const OBJECTIVE_STATUS_MAP: Record<string, { color: string; label: string }> = {
	not_started: { color: '#6b7280', label: 'Not Started' },
	in_progress: { color: '#f59e0b', label: 'In Progress' },
	completed: { color: '#22c55e', label: 'Completed' },
	on_hold: { color: '#3b82f6', label: 'On Hold' },
	cancelled: { color: '#ef4444', label: 'Cancelled' }
};

export function getObjectiveStatusColor(status: string): string {
	return OBJECTIVE_STATUS_MAP[status]?.color ?? '#6b7280';
}

export function getObjectiveStatusLabel(status: string): string {
	return OBJECTIVE_STATUS_MAP[status]?.label ?? status;
}

// 日付フォーマット
export function formatDate(dateStr?: string): string {
	if (!dateStr || dateStr.length < 10) return '-';
	return dateStr.slice(0, 10);
}

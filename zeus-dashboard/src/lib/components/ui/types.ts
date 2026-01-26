// UI コンポーネント用型定義

/**
 * Tooltip に表示するエンティティ情報
 */
export interface TooltipEntity {
	id: string;
	title: string;
	type: 'vision' | 'objective' | 'deliverable' | 'task';
	status: string;
	progress: number;
	lastUpdate?: string;
}

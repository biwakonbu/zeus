// デザイントークン - Factorio 風 UI システム
// 議論結果（round: 20260121-174500_wbsdesign）に基づく

export const tokens = {
	colors: {
		accent: {
			primary: '#ff9533',
			glow: 'rgba(255, 149, 51, 0.15)'
		},
		status: {
			completed: '#44cc44',
			inProgress: '#4488ff',
			notStarted: '#666666',
			blocked: '#ee4444',
			warning: '#ffcc00'
		},
		background: {
			dark: '#1a1a1a',
			mid: '#2d2d2d',
			elevated: '#3a3a3a'
		},
		border: {
			subtle: '#333333',
			mid: '#4a4a4a',
			strong: '#5a5a5a'
		}
	},
	spacing: {
		xs: '4px',
		sm: '8px',
		md: '12px',
		lg: '16px',
		xl: '24px'
	},
	radius: {
		sm: '2px',
		md: '4px', // 基準値（議論結果: 角丸4px）
		lg: '8px'
	},
	shadows: {
		tooltip: '0 4px 20px rgba(0, 0, 0, 0.5)',
		card: '0 2px 8px rgba(0, 0, 0, 0.3)',
		glow: '0 0 15px rgba(255, 149, 51, 0.3)',
		inset: 'inset 0 1px 0 rgba(255, 255, 255, 0.03)'
	},
	timing: {
		tooltipDelay: 500,
		transitionFast: 100,
		transitionNormal: 200,
		transitionSlow: 300
	},
	tooltip: {
		width: 320,
		height: 220,
		offset: 16
	}
} as const;

export type Tokens = typeof tokens;
export type StatusColor = keyof typeof tokens.colors.status;
export type EntityType = 'vision' | 'objective' | 'deliverable' | 'activity' | 'usecase';

// ステータスから色を取得するヘルパー
export function getStatusColor(status: string): string {
	switch (status) {
		case 'completed':
			return tokens.colors.status.completed;
		case 'in_progress':
			return tokens.colors.status.inProgress;
		case 'blocked':
			return tokens.colors.status.blocked;
		case 'not_started':
			return tokens.colors.status.notStarted;
		default:
			return tokens.colors.status.notStarted;
	}
}

// エンティティタイプからアイコン名を取得するヘルパー
export function getEntityIcon(type: EntityType): string {
	switch (type) {
		case 'vision':
			return 'Target';
		case 'objective':
			return 'Flag';
		case 'deliverable':
			return 'Package';
		case 'activity':
			return 'CheckSquare';
		case 'usecase':
			return 'Users';
		default:
			return 'Circle';
	}
}

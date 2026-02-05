import { http, HttpResponse } from 'msw';
import type {
	GraphNode,
	StatusResponse,
	Vision,
	VisionResponse,
	Objective,
	ObjectivesResponse,
	Deliverable,
	DeliverablesResponse,
	Consideration,
	ConsiderationsResponse,
	Decision,
	DecisionsResponse,
	Problem,
	ProblemsResponse,
	Risk,
	RisksResponse,
	Assumption,
	AssumptionsResponse,
	Constraint,
	ConstraintsResponse,
	QualityItem,
	QualityResponse
} from '../src/lib/types/api';

// モックデータ: GraphNode
const mockGraphNodes: GraphNode[] = [
	{
		id: 'task-1',
		title: 'プロジェクト初期化',
		node_type: 'task',
		status: 'completed',
		priority: 'high',
		assignee: 'alice',
		dependencies: []
	},
	{
		id: 'task-2',
		title: 'データベース設計',
		node_type: 'task',
		status: 'completed',
		priority: 'high',
		assignee: 'bob',
		dependencies: ['task-1']
	},
	{
		id: 'task-3',
		title: 'API 実装',
		node_type: 'task',
		status: 'in_progress',
		priority: 'high',
		assignee: 'alice',
		dependencies: ['task-2']
	},
	{
		id: 'task-4',
		title: 'フロントエンド実装',
		node_type: 'task',
		status: 'in_progress',
		priority: 'medium',
		assignee: 'charlie',
		dependencies: ['task-2']
	},
	{
		id: 'task-5',
		title: 'テスト作成',
		node_type: 'task',
		status: 'pending',
		priority: 'medium',
		assignee: 'bob',
		dependencies: ['task-3', 'task-4']
	},
	{
		id: 'task-6',
		title: 'デプロイメント設定',
		node_type: 'task',
		status: 'blocked',
		priority: 'low',
		assignee: 'alice',
		dependencies: ['task-5']
	}
];

// モックデータ: ステータス
const mockStatus: StatusResponse = {
	project: {
		id: 'zeus-demo',
		name: 'Zeus Demo Project',
		description: 'Storybook デモ用プロジェクト',
		start_date: '2024-01-01'
	},
	state: {
		health: 'good',
		summary: {
			total_activities: 6,
			completed: 2,
			in_progress: 2,
			pending: 2
		}
	},
	pending_approvals: 1
};

// =============================================================================
// 10概念モデル モックデータ
// =============================================================================

// モックデータ: Vision
const mockVision: Vision = {
	title: 'プロジェクト管理の革新',
	statement:
		'Zeus は AI 駆動のプロジェクト管理ツールとして、チームの生産性を最大化し、プロジェクトの成功率を向上させる',
	created_at: '2024-01-01T00:00:00Z',
	updated_at: '2024-01-01T00:00:00Z'
};

// モックデータ: Objectives
const mockObjectives: Objective[] = [
	{
		id: 'obj-001',
		title: 'Phase 1: MVP 開発',
		description: 'コア機能の実装と基盤構築',
		status: 'completed',
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-31T00:00:00Z'
	},
	{
		id: 'obj-002',
		title: 'Phase 2: 標準機能',
		description: '承認システムとスナップショット機能',
		status: 'in_progress',
		parent_id: 'obj-001',
		created_at: '2024-02-01T00:00:00Z',
		updated_at: '2024-02-15T00:00:00Z'
	},
	{
		id: 'obj-003',
		title: 'Phase 3: AI 統合',
		description: 'Claude Code 連携と分析機能',
		status: 'not_started',
		created_at: '2024-02-15T00:00:00Z',
		updated_at: '2024-02-15T00:00:00Z'
	}
];

// モックデータ: Deliverables
const mockDeliverables: Deliverable[] = [
	{
		id: 'del-001',
		title: 'システム設計書',
		description: 'アーキテクチャと技術スタックの詳細設計',
		format: 'document',
		objective_id: 'obj-001',
		status: 'approved',
		due_date: '2024-01-15',
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-15T00:00:00Z'
	},
	{
		id: 'del-002',
		title: 'CLI ツール',
		description: 'Zeus CLI バイナリ',
		format: 'code',
		objective_id: 'obj-001',
		status: 'delivered',
		due_date: '2024-01-31',
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-31T00:00:00Z'
	},
	{
		id: 'del-003',
		title: 'ダッシュボード',
		description: 'Web ダッシュボード（SvelteKit）',
		format: 'code',
		objective_id: 'obj-002',
		status: 'in_review',
		due_date: '2024-02-28',
		created_at: '2024-02-01T00:00:00Z',
		updated_at: '2024-02-20T00:00:00Z'
	}
];

// モックデータ: Considerations
const mockConsiderations: Consideration[] = [
	{
		id: 'con-001',
		title: 'フロントエンドフレームワーク選定',
		description: 'ダッシュボードに使用するフレームワークの検討',
		status: 'decided',
		objective_id: 'obj-002',
		options: [
			{
				id: 'opt-1',
				title: 'SvelteKit',
				description: '軽量で高速、学習コスト低',
				pros: ['バンドルサイズ小', '高パフォーマンス', 'TypeScript サポート'],
				cons: ['エコシステムが小さい']
			},
			{
				id: 'opt-2',
				title: 'React',
				description: '最大のエコシステム',
				pros: ['豊富なライブラリ', '大規模コミュニティ'],
				cons: ['バンドルサイズ大', 'ボイラープレート多い']
			}
		],
		decision_id: 'dec-001',
		created_at: '2024-01-10T00:00:00Z',
		updated_at: '2024-01-15T00:00:00Z'
	},
	{
		id: 'con-002',
		title: 'グラフ描画ライブラリ選定',
		description: 'タスクグラフの描画に使用するライブラリ',
		status: 'open',
		objective_id: 'obj-002',
		options: [
			{
				id: 'opt-1',
				title: 'PixiJS',
				description: 'WebGL ベースの高パフォーマンス描画',
				pros: ['高パフォーマンス', '大量ノード対応'],
				cons: ['学習コスト高']
			},
			{
				id: 'opt-2',
				title: 'D3.js',
				description: 'データ可視化の定番',
				pros: ['豊富な機能', 'SVG 対応'],
				cons: ['大量ノードで重い']
			}
		],
		due_date: '2024-02-20',
		created_at: '2024-02-01T00:00:00Z',
		updated_at: '2024-02-10T00:00:00Z'
	}
];

// モックデータ: Decisions
const mockDecisions: Decision[] = [
	{
		id: 'dec-001',
		title: 'SvelteKit 採用決定',
		consideration_id: 'con-001',
		selected_option: {
			id: 'opt-1',
			title: 'SvelteKit'
		},
		rationale: 'パフォーマンスと開発効率のバランスが最も良い',
		decided_at: '2024-01-15T00:00:00Z',
		created_at: '2024-01-15T00:00:00Z'
	}
];

// モックデータ: Problems
const mockProblems: Problem[] = [
	{
		id: 'prob-001',
		title: 'SSE 接続が不安定',
		description: 'ネットワーク切断時に自動再接続が失敗することがある',
		severity: 'medium',
		status: 'investigating',
		objective_id: 'obj-002',
		created_at: '2024-02-10T00:00:00Z',
		updated_at: '2024-02-15T00:00:00Z'
	},
	{
		id: 'prob-002',
		title: '大量ノードでのパフォーマンス低下',
		description: '100+ ノードで FPS が低下する',
		severity: 'high',
		status: 'open',
		deliverable_id: 'del-003',
		created_at: '2024-02-18T00:00:00Z',
		updated_at: '2024-02-18T00:00:00Z'
	}
];

// モックデータ: Risks
const mockRisks: Risk[] = [
	{
		id: 'risk-001',
		title: '外部 API 依存',
		description: 'Claude API の可用性に依存',
		probability: 'low',
		impact: 'high',
		score: 3,
		status: 'mitigating',
		objective_id: 'obj-003',
		mitigation: 'フォールバック機構の実装',
		created_at: '2024-01-05T00:00:00Z',
		updated_at: '2024-02-01T00:00:00Z'
	},
	{
		id: 'risk-002',
		title: 'スケジュール遅延',
		description: 'Phase 2 の遅延が Phase 3 に影響',
		probability: 'medium',
		impact: 'medium',
		score: 4,
		status: 'identified',
		objective_id: 'obj-002',
		created_at: '2024-02-10T00:00:00Z',
		updated_at: '2024-02-10T00:00:00Z'
	}
];

// モックデータ: Assumptions
const mockAssumptions: Assumption[] = [
	{
		id: 'assum-001',
		title: 'ユーザー数 1000 人以下',
		description: '初期リリース時のターゲットユーザー数',
		status: 'verified',
		objective_id: 'obj-001',
		verified_at: '2024-01-20T00:00:00Z',
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-20T00:00:00Z'
	},
	{
		id: 'assum-002',
		title: 'Go 1.21+ 環境',
		description: 'ユーザーの実行環境に Go 1.21 以上がインストールされている',
		status: 'unverified',
		objective_id: 'obj-001',
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-01T00:00:00Z'
	}
];

// モックデータ: Constraints
const mockConstraints: Constraint[] = [
	{
		id: 'const-001',
		title: '外部 DB 不使用',
		description: 'ファイルベースのデータストアのみ使用',
		category: 'technical',
		non_negotiable: true,
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-01T00:00:00Z'
	},
	{
		id: 'const-002',
		title: 'OSS ライセンス',
		description: 'MIT ライセンスで公開',
		category: 'legal',
		non_negotiable: true,
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-01T00:00:00Z'
	},
	{
		id: 'const-003',
		title: '開発期間 3ヶ月',
		description: 'MVP リリースまでの期限',
		category: 'time',
		non_negotiable: false,
		created_at: '2024-01-01T00:00:00Z',
		updated_at: '2024-01-01T00:00:00Z'
	}
];

// モックデータ: Quality
const mockQuality: QualityItem[] = [
	{
		id: 'qual-001',
		title: 'コードカバレッジ',
		description: 'ユニットテストのカバレッジ目標',
		deliverable_id: 'del-002',
		metric: {
			name: 'coverage',
			target: 80,
			current: 75,
			unit: '%'
		},
		status: 'failing',
		created_at: '2024-01-15T00:00:00Z',
		updated_at: '2024-02-01T00:00:00Z'
	},
	{
		id: 'qual-002',
		title: 'パフォーマンス基準',
		description: 'CLI レスポンスタイム',
		deliverable_id: 'del-002',
		metric: {
			name: 'response_time',
			target: 100,
			current: 50,
			unit: 'ms'
		},
		status: 'passing',
		created_at: '2024-01-15T00:00:00Z',
		updated_at: '2024-02-01T00:00:00Z'
	},
	{
		id: 'qual-003',
		title: 'セキュリティ監査',
		description: 'セキュリティレビュー完了',
		deliverable_id: 'del-002',
		gate: {
			name: 'security_audit',
			passed: true,
			checked_at: '2024-01-30T00:00:00Z'
		},
		status: 'passing',
		created_at: '2024-01-20T00:00:00Z',
		updated_at: '2024-01-30T00:00:00Z'
	}
];

// MSW ハンドラー
export const handlers = [
	// ステータス
	http.get('/api/status', () => {
		return HttpResponse.json(mockStatus);
	}),

	// グラフ（Mermaid）
	http.get('/api/graph', () => {
		return HttpResponse.json({
			mermaid: `graph TD
    task-1[プロジェクト初期化] --> task-2[データベース設計]
    task-2 --> task-3[API 実装]
    task-2 --> task-4[フロントエンド実装]
    task-3 --> task-5[テスト作成]
    task-4 --> task-5
    task-5 --> task-6[デプロイメント設定]`,
			stats: {
				total_nodes: 6,
				with_dependencies: 5,
				isolated_count: 0,
				cycle_count: 0,
				max_depth: 4
			},
			cycles: [],
			isolated: []
		});
	}),

	// ==========================================================================
	// 10概念モデル API ハンドラー
	// ==========================================================================

	// Vision
	http.get('/api/vision', () => {
		const response: VisionResponse = { vision: mockVision };
		return HttpResponse.json(response);
	}),

	// Objectives
	http.get('/api/objectives', () => {
		const response: ObjectivesResponse = {
			objectives: mockObjectives,
			total: mockObjectives.length
		};
		return HttpResponse.json(response);
	}),

	// Deliverables
	http.get('/api/deliverables', () => {
		const response: DeliverablesResponse = {
			deliverables: mockDeliverables,
			total: mockDeliverables.length
		};
		return HttpResponse.json(response);
	}),

	// Considerations
	http.get('/api/considerations', () => {
		const response: ConsiderationsResponse = {
			considerations: mockConsiderations,
			total: mockConsiderations.length
		};
		return HttpResponse.json(response);
	}),

	// Decisions
	http.get('/api/decisions', () => {
		const response: DecisionsResponse = {
			decisions: mockDecisions,
			total: mockDecisions.length
		};
		return HttpResponse.json(response);
	}),

	// Problems
	http.get('/api/problems', () => {
		const response: ProblemsResponse = {
			problems: mockProblems,
			total: mockProblems.length
		};
		return HttpResponse.json(response);
	}),

	// Risks
	http.get('/api/risks', () => {
		const response: RisksResponse = {
			risks: mockRisks,
			total: mockRisks.length
		};
		return HttpResponse.json(response);
	}),

	// Assumptions
	http.get('/api/assumptions', () => {
		const response: AssumptionsResponse = {
			assumptions: mockAssumptions,
			total: mockAssumptions.length
		};
		return HttpResponse.json(response);
	}),

	// Constraints
	http.get('/api/constraints', () => {
		const response: ConstraintsResponse = {
			constraints: mockConstraints,
			total: mockConstraints.length
		};
		return HttpResponse.json(response);
	}),

	// Quality
	http.get('/api/quality', () => {
		const response: QualityResponse = {
			quality_items: mockQuality,
			total: mockQuality.length
		};
		return HttpResponse.json(response);
	})
];

// エクスポート: 既存 + 10概念モデル
export {
	mockGraphNodes,
	mockWBS,
	mockTimeline,
	mockStatus,
	mockVision,
	mockObjectives,
	mockDeliverables,
	mockConsiderations,
	mockDecisions,
	mockProblems,
	mockRisks,
	mockAssumptions,
	mockConstraints,
	mockQuality
};

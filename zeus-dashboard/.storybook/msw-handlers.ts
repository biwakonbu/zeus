import { http, HttpResponse } from 'msw';
import type {
	TaskItem,
	TasksResponse,
	WBSNode,
	WBSResponse,
	TimelineItem,
	TimelineResponse,
	DownstreamResponse,
	StatusResponse
} from '../src/lib/types/api';

// モックデータ: タスク
const mockTasks: TaskItem[] = [
	{
		id: 'task-1',
		title: 'プロジェクト初期化',
		status: 'completed',
		priority: 'high',
		assignee: 'alice',
		dependencies: [],
		progress: 100,
		wbs_code: '1',
		start_date: '2024-01-01',
		due_date: '2024-01-05'
	},
	{
		id: 'task-2',
		title: 'データベース設計',
		status: 'completed',
		priority: 'high',
		assignee: 'bob',
		dependencies: ['task-1'],
		progress: 100,
		wbs_code: '2',
		start_date: '2024-01-06',
		due_date: '2024-01-10'
	},
	{
		id: 'task-3',
		title: 'API 実装',
		status: 'in_progress',
		priority: 'high',
		assignee: 'alice',
		dependencies: ['task-2'],
		progress: 60,
		wbs_code: '3',
		start_date: '2024-01-11',
		due_date: '2024-01-20'
	},
	{
		id: 'task-4',
		title: 'フロントエンド実装',
		status: 'in_progress',
		priority: 'medium',
		assignee: 'charlie',
		dependencies: ['task-2'],
		progress: 40,
		wbs_code: '4',
		start_date: '2024-01-11',
		due_date: '2024-01-25'
	},
	{
		id: 'task-5',
		title: 'テスト作成',
		status: 'pending',
		priority: 'medium',
		assignee: 'bob',
		dependencies: ['task-3', 'task-4'],
		progress: 0,
		wbs_code: '5',
		start_date: '2024-01-26',
		due_date: '2024-02-01'
	},
	{
		id: 'task-6',
		title: 'デプロイメント設定',
		status: 'blocked',
		priority: 'low',
		assignee: 'alice',
		dependencies: ['task-5'],
		progress: 0,
		wbs_code: '6',
		start_date: '2024-02-02',
		due_date: '2024-02-05'
	}
];

// モックデータ: WBS
const mockWBS: WBSResponse = {
	roots: [
		{
			id: 'task-1',
			title: 'プロジェクト初期化',
			wbs_code: '1',
			status: 'completed',
			progress: 100,
			priority: 'high',
			assignee: 'alice',
			depth: 0,
			children: [
				{
					id: 'task-2',
					title: 'データベース設計',
					wbs_code: '1.1',
					status: 'completed',
					progress: 100,
					priority: 'high',
					assignee: 'bob',
					depth: 1,
					children: [
						{
							id: 'task-3',
							title: 'API 実装',
							wbs_code: '1.1.1',
							status: 'in_progress',
							progress: 60,
							priority: 'high',
							assignee: 'alice',
							depth: 2
						},
						{
							id: 'task-4',
							title: 'フロントエンド実装',
							wbs_code: '1.1.2',
							status: 'in_progress',
							progress: 40,
							priority: 'medium',
							assignee: 'charlie',
							depth: 2
						}
					]
				}
			]
		}
	],
	max_depth: 3,
	stats: {
		total_nodes: 6,
		root_count: 1,
		leaf_count: 4,
		max_depth: 3,
		avg_progress: 50,
		completed_pct: 33
	}
};

// モックデータ: タイムライン
const mockTimeline: TimelineResponse = {
	items: mockTasks.map((task) => ({
		task_id: task.id,
		title: task.title,
		start_date: task.start_date || '2024-01-01',
		end_date: task.due_date || '2024-01-10',
		progress: task.progress,
		status: task.status,
		priority: task.priority,
		assignee: task.assignee,
		is_on_critical_path: ['task-1', 'task-2', 'task-3', 'task-5', 'task-6'].includes(task.id),
		slack: task.status === 'completed' ? null : Math.floor(Math.random() * 5),
		dependencies: task.dependencies
	})),
	critical_path: ['task-1', 'task-2', 'task-3', 'task-5', 'task-6'],
	project_start: '2024-01-01',
	project_end: '2024-02-05',
	total_duration: 36,
	stats: {
		total_tasks: 6,
		tasks_with_dates: 6,
		on_critical_path: 5,
		average_slack: 2.5,
		overdue_tasks: 0,
		completed_on_time: 2
	}
};

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
			total_tasks: 6,
			completed: 2,
			in_progress: 2,
			pending: 2
		}
	},
	pending_approvals: 1
};

// MSW ハンドラー
export const handlers = [
	// タスク一覧
	http.get('/api/tasks', () => {
		const response: TasksResponse = {
			tasks: mockTasks,
			total: mockTasks.length
		};
		return HttpResponse.json(response);
	}),

	// WBS
	http.get('/api/wbs', () => {
		return HttpResponse.json(mockWBS);
	}),

	// タイムライン
	http.get('/api/timeline', () => {
		return HttpResponse.json(mockTimeline);
	}),

	// 下流タスク
	http.get('/api/downstream', ({ request }) => {
		const url = new URL(request.url);
		const taskId = url.searchParams.get('task_id') || 'task-1';

		// 簡易的な下流・上流計算
		const task = mockTasks.find((t) => t.id === taskId);
		const downstream = mockTasks.filter((t) => t.dependencies.includes(taskId)).map((t) => t.id);
		const upstream = task?.dependencies || [];

		const response: DownstreamResponse = {
			task_id: taskId,
			downstream,
			upstream,
			count: downstream.length
		};
		return HttpResponse.json(response);
	}),

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

	// 予測
	http.get('/api/predict', () => {
		return HttpResponse.json({
			completion: {
				remaining_tasks: 4,
				average_velocity: 1.5,
				estimated_date: '2024-02-10',
				confidence_level: 0.75,
				margin_days: 5,
				has_sufficient_data: true
			},
			risk: {
				overall_level: 'medium',
				factors: [
					{
						name: 'ブロッキングタスク',
						description: '1件のタスクがブロックされています',
						impact: 0.3
					}
				],
				score: 0.4
			},
			velocity: {
				last_7_days: 2,
				last_14_days: 3,
				last_30_days: 5,
				weekly_average: 1.5,
				trend: 'stable',
				data_points: 10
			}
		});
	})
];

export { mockTasks, mockWBS, mockTimeline, mockStatus };

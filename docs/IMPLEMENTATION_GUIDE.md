# Zeus 実装ガイド

## 1. はじめに

このガイドでは、Zeus システムをClaude Code Pluginとして実装する手順を説明します。

## 2. プロジェクト構造

### 2.1 ディレクトリ構成
```
zeus-plugin/
├── plugin.json              # プラグイン定義
├── package.json             # Node.js依存関係
├── README.md                # プラグイン説明
│
├── commands/                # スラッシュコマンド
│   ├── init.md              # /zeus-init
│   ├── status.md            # /zeus-status
│   └── suggest.md           # /zeus-suggest
│
├── skills/                  # Agent Skills
│   ├── project-scan/        # プロジェクトスキャン
│   │   ├── skill.js
│   │   └── SKILL.md
│   ├── task-suggest/        # タスク提案
│   │   ├── skill.js
│   │   └── SKILL.md
│   ├── risk-analysis/       # リスク分析
│   │   ├── skill.js
│   │   └── SKILL.md
│   └── timeline-optimize/   # タイムライン最適化
│       ├── skill.js
│       └── SKILL.md
│
├── lib/                     # 共通ライブラリ
│   ├── zeus-core.js         # コア機能
│   ├── file-manager.js      # ファイル操作
│   ├── yaml-parser.js       # YAML処理
│   ├── state-manager.js     # 状態管理
│   ├── approval-manager.js  # 承認管理
│   └── ai-interface.js      # AI連携
│
├── templates/               # テンプレート
│   ├── zeus.yaml.template
│   ├── task.yaml.template
│   └── report.html.template
│
└── tests/                   # テストコード
    ├── unit/
    └── integration/
```

## 3. プラグイン定義

### 3.1 plugin.json
```json
{
  "name": "zeus",
  "version": "1.0.0",
  "description": "AI-driven project management system with god's eye view",
  "author": "Zeus Development Team",
  "keywords": ["project-management", "ai", "wbs", "timeline"],
  "repository": "https://github.com/biwakonbu/zeus",
  "license": "MIT",
  "commands": [
    {
      "name": "zeus-init",
      "description": "Initialize Zeus project",
      "path": "commands/init.md"
    },
    {
      "name": "zeus-status",
      "description": "Show project status overview",
      "path": "commands/status.md"
    },
    {
      "name": "zeus-suggest",
      "description": "Get AI suggestions",
      "path": "commands/suggest.md"
    }
  ],
  "skills": [
    {
      "name": "project-scan",
      "description": "Scan and analyze project structure",
      "path": "skills/project-scan"
    },
    {
      "name": "task-suggest",
      "description": "Suggest task breakdown and priorities",
      "path": "skills/task-suggest"
    },
    {
      "name": "risk-analysis",
      "description": "Analyze project risks",
      "path": "skills/risk-analysis"
    },
    {
      "name": "timeline-optimize",
      "description": "Optimize project timeline",
      "path": "skills/timeline-optimize"
    }
  ]
}
```

## 4. コア実装

### 4.1 zeus-core.js
```javascript
// lib/zeus-core.js
import { FileManager } from './file-manager.js';
import { YamlParser } from './yaml-parser.js';
import { StateManager } from './state-manager.js';
import { ApprovalManager } from './approval-manager.js';

export class ZeusCore {
  constructor(projectPath = '.') {
    this.projectPath = projectPath;
    this.zeusPath = `${projectPath}/.zeus`;
    this.fileManager = new FileManager(this.zeusPath);
    this.yamlParser = new YamlParser();
    this.stateManager = new StateManager(this.fileManager, this.yamlParser);
    this.approvalManager = new ApprovalManager(this.fileManager, this.yamlParser);
  }

  async init(level = 'simple') {
    // ディレクトリ構造を作成
    const dirs = this.getDirectoryStructure(level);
    for (const dir of dirs) {
      await this.fileManager.ensureDir(dir);
    }

    // zeus.yamlを生成
    const zeusConfig = this.generateInitialConfig();
    await this.fileManager.writeYaml('zeus.yaml', zeusConfig);

    // 初期状態を記録
    await this.stateManager.createSnapshot('initial');

    return { success: true, level, path: this.zeusPath };
  }

  async status(options = {}) {
    const config = await this.fileManager.readYaml('zeus.yaml');
    const state = await this.stateManager.getCurrentState();
    const pending = await this.approvalManager.getPending();

    if (options.detail) {
      return this.generateDetailedStatus(config, state, pending);
    }

    return this.generateQuickStatus(config, state, pending);
  }

  async suggest() {
    const state = await this.stateManager.getCurrentState();
    const suggestions = await this.generateSuggestions(state);

    // 各提案に承認レベルを設定
    for (const suggestion of suggestions) {
      suggestion.approvalLevel = this.determineApprovalLevel(suggestion);
    }

    return suggestions;
  }

  async applyApproval(id, approved, reason = null) {
    const approval = await this.approvalManager.get(id);
    if (!approval) {
      throw new Error(`Approval ${id} not found`);
    }

    if (approved) {
      await this.executeSuggestion(approval.suggestion);
      await this.approvalManager.markApproved(id);
    } else {
      await this.approvalManager.markRejected(id, reason);
    }

    // スナップショットを作成
    await this.stateManager.createSnapshot(`approval-${id}`);

    return { success: true, id, approved };
  }

  // プライベートメソッド
  getDirectoryStructure(level) {
    const dirs = {
      simple: ['.'],
      standard: [
        'config', 'tasks', 'state', 'entities',
        'approvals/pending', 'approvals/approved', 'approvals/rejected',
        'logs', 'analytics'
      ],
      advanced: [
        'config', 'tasks', 'state', 'entities',
        'approvals/pending', 'approvals/approved', 'approvals/rejected',
        'logs', 'analytics', 'graph', 'views', '.local'
      ]
    };

    return dirs[level] || dirs.simple;
  }

  generateInitialConfig() {
    return {
      version: '1.0',
      project: {
        id: `zeus-${Date.now()}`,
        name: 'New Zeus Project',
        description: 'Project managed by Zeus',
        start_date: new Date().toISOString().split('T')[0]
      },
      objectives: [],
      settings: {
        automation_level: 'standard',
        approval_mode: 'default',
        ai_provider: 'claude-code'
      }
    };
  }

  determineApprovalLevel(suggestion) {
    // ルールベースで承認レベルを決定
    if (suggestion.type === 'read' || suggestion.type === 'report') {
      return 'auto';
    }
    if (suggestion.impact === 'low' && suggestion.risk === 'low') {
      return 'notify';
    }
    return 'approve';
  }
}
```

### 4.2 状態管理
```javascript
// lib/state-manager.js
export class StateManager {
  constructor(fileManager, yamlParser) {
    this.fileManager = fileManager;
    this.yamlParser = yamlParser;
  }

  async getCurrentState() {
    try {
      return await this.fileManager.readYaml('state/current.yaml');
    } catch (e) {
      return this.getEmptyState();
    }
  }

  async createSnapshot(label = null) {
    const state = await this.calculateCurrentState();
    const timestamp = new Date().toISOString();
    const filename = `state/snapshots/${timestamp.replace(/[:.]/g, '-')}.yaml`;

    const snapshot = {
      timestamp,
      label,
      state
    };

    await this.fileManager.writeYaml(filename, snapshot);
    await this.fileManager.writeYaml('state/current.yaml', state);

    return snapshot;
  }

  async calculateCurrentState() {
    const tasks = await this.getTaskStats();
    const health = this.calculateHealth(tasks);
    const risks = await this.identifyRisks();

    return {
      timestamp: new Date().toISOString(),
      summary: tasks,
      health,
      risks
    };
  }

  async getTaskStats() {
    const active = await this.fileManager.readYaml('tasks/active.yaml') || { tasks: [] };
    const backlog = await this.fileManager.readYaml('tasks/backlog.yaml') || { tasks: [] };

    const allTasks = [...active.tasks, ...backlog.tasks];

    return {
      total_tasks: allTasks.length,
      completed: allTasks.filter(t => t.status === 'completed').length,
      in_progress: allTasks.filter(t => t.status === 'in_progress').length,
      pending: allTasks.filter(t => t.status === 'pending').length
    };
  }

  calculateHealth(taskStats) {
    const progress = taskStats.completed / (taskStats.total_tasks || 1);
    if (progress < 0.3) return 'poor';
    if (progress < 0.7) return 'fair';
    return 'good';
  }

  getEmptyState() {
    return {
      timestamp: new Date().toISOString(),
      summary: {
        total_tasks: 0,
        completed: 0,
        in_progress: 0,
        pending: 0
      },
      health: 'unknown',
      risks: []
    };
  }
}
```

## 5. スキル実装

### 5.1 project-scan スキル
```javascript
// skills/project-scan/skill.js
export async function projectScan(context) {
  const { fileSystem, yaml } = context.tools;

  // プロジェクトファイルを探索
  const projectFiles = await findProjectFiles(fileSystem);

  // プロジェクト構造を分析
  const structure = await analyzeProjectStructure(projectFiles);

  // 既存のタスク管理ツールをチェック
  const existingTools = await detectExistingTools(projectFiles);

  // zeus.yamlの初期構造を生成
  const zeusConfig = generateZeusConfig(structure, existingTools);

  // タスクの初期セットを提案
  const initialTasks = suggestInitialTasks(structure);

  return {
    projectStructure: structure,
    existingTools,
    suggestedConfig: zeusConfig,
    suggestedTasks: initialTasks
  };
}

async function findProjectFiles(fs) {
  const patterns = [
    'package.json',
    'requirements.txt',
    'Gemfile',
    'Cargo.toml',
    'go.mod',
    'README*',
    'Makefile',
    '.git/config'
  ];

  const files = [];
  for (const pattern of patterns) {
    const found = await fs.glob(pattern);
    files.push(...found);
  }

  return files;
}

function generateZeusConfig(structure, existingTools) {
  return {
    version: '1.0',
    project: {
      id: `zeus-${Date.now()}`,
      name: 'Project',
      description: `${structure.type} project using ${structure.languages.join(', ')}`,
      tech_stack: {
        languages: structure.languages,
        frameworks: structure.frameworks
      },
      integration: {
        existing_tools: existingTools
      }
    }
  };
}
```

### 5.2 task-suggest スキル
```javascript
// skills/task-suggest/skill.js
export async function taskSuggest(context) {
  const { zeus, ai } = context.tools;

  // 現在の状態を取得
  const state = await zeus.getState();
  const objectives = await zeus.getObjectives();

  // タスクブレークダウン
  const breakdown = await generateTaskBreakdown(objectives, state, ai);

  // 優先順位付け
  const prioritized = prioritizeTasks(breakdown);

  // 依存関係の特定
  const withDependencies = identifyDependencies(prioritized);

  // タイムライン推定
  const timeline = estimateTimeline(withDependencies);

  return {
    tasks: withDependencies,
    timeline,
    criticalPath: calculateCriticalPath(withDependencies)
  };
}

function prioritizeTasks(tasks) {
  return tasks.sort((a, b) => {
    const scoreA = calculatePriorityScore(a);
    const scoreB = calculatePriorityScore(b);
    return scoreB - scoreA;
  });
}

function calculatePriorityScore(task) {
  let score = 0;
  score += task.business_value || 0;
  if (task.mitigates_risk) score += 20;
  if (task.is_blocker) score += 30;
  if (task.urgent) score += 25;
  return score;
}
```

## 6. エラー処理と復旧

### 6.1 doctor コマンド実装
```javascript
// lib/doctor.js
export class ZeusDoctor {
  constructor(zeusCore) {
    this.core = zeusCore;
    this.checks = [
      this.checkConfigExists,
      this.checkYamlSyntax,
      this.checkStateIntegrity,
      this.checkTaskConsistency,
      this.checkApprovalQueue,
      this.checkBackupHealth
    ];
  }

  async diagnose() {
    const results = [];

    for (const check of this.checks) {
      try {
        const result = await check.call(this);
        results.push(result);
      } catch (error) {
        results.push({
          check: check.name,
          status: 'error',
          message: error.message,
          fixable: false
        });
      }
    }

    return {
      overall: this.calculateOverallHealth(results),
      checks: results,
      fixableCount: results.filter(r => r.fixable).length
    };
  }

  async fix(dryRun = false) {
    const diagnosis = await this.diagnose();
    const fixes = [];

    for (const check of diagnosis.checks) {
      if (check.status === 'fail' && check.fixable) {
        if (dryRun) {
          fixes.push({
            action: check.fix.description,
            would_execute: true
          });
        } else {
          await check.fix.action();
          fixes.push({
            action: check.fix.description,
            executed: true
          });
        }
      }
    }

    return { fixes, dryRun };
  }

  async checkConfigExists() {
    const exists = await this.core.fileManager.exists('zeus.yaml');
    return {
      check: 'config_exists',
      status: exists ? 'pass' : 'fail',
      message: exists ? 'Configuration file found' : 'zeus.yaml not found',
      fixable: true,
      fix: {
        description: 'Create default zeus.yaml',
        action: () => this.core.init('simple')
      }
    };
  }

  calculateOverallHealth(results) {
    const failed = results.filter(r => r.status === 'fail').length;
    const warned = results.filter(r => r.status === 'warn').length;

    if (failed > 0) return 'unhealthy';
    if (warned > 2) return 'degraded';
    if (warned > 0) return 'fair';
    return 'healthy';
  }
}
```

## 7. テスト戦略

### 7.1 ユニットテスト
```javascript
// tests/unit/zeus-core.test.js
import { describe, it, expect, beforeEach } from 'vitest';
import { ZeusCore } from '../../lib/zeus-core.js';

describe('ZeusCore', () => {
  let zeus;

  beforeEach(() => {
    zeus = new ZeusCore('.');
  });

  describe('init', () => {
    it('should create simple structure by default', async () => {
      const result = await zeus.init();
      expect(result.success).toBe(true);
      expect(result.level).toBe('simple');
    });
  });

  describe('status', () => {
    it('should return quick status by default', async () => {
      const status = await zeus.status();
      expect(status.summary).toBeDefined();
    });
  });
});
```

## 8. デプロイメント

### 8.1 パッケージング
```bash
# package.json scripts
{
  "scripts": {
    "build": "tsc && npm run bundle",
    "test": "vitest run",
    "lint": "eslint lib/ skills/",
    "publish:marketplace": "claude-plugin publish"
  }
}
```

### 8.2 Claude Code Marketplace 公開
```bash
# プラグインの検証
claude plugin validate ./zeus-plugin

# マーケットプレイスに公開
claude plugin publish ./zeus-plugin
```

## 9. 次のステップ

1. **Phase 1 実装**: 基本コマンドとcore機能
2. **テスト作成**: ユニット・統合テストの完成
3. **ドキュメント**: APIリファレンスの作成
4. **Phase 2 計画**: 高度な機能の設計詳細化

---

*Zeus Implementation Guide v1.0*
*作成日: 2026-01-14*
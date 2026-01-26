// JoinNode - UML アクティビティ図の並列合流ノード（太い横線）
import type { ActivityNodeItem } from '$lib/types/api';
import { SyncBarNode } from './SyncBarNode';

/**
 * JoinNode - 並列合流ノード
 *
 * UML 表記: 太い横線（同期バー）
 * 複数の並列フローが1つの制御フローに合流する点を表す
 *
 * SyncBarNode を継承（ForkNode と共通の描画ロジック）
 */
export class JoinNode extends SyncBarNode {
	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);
	}
}

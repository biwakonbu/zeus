// ForkNode - UML アクティビティ図の並列分岐ノード（太い横線）
import type { ActivityNodeItem } from '$lib/types/api';
import { SyncBarNode } from './SyncBarNode';

/**
 * ForkNode - 並列分岐ノード
 *
 * UML 表記: 太い横線（同期バー）
 * 1つの制御フローが複数の並列フローに分岐する点を表す
 *
 * SyncBarNode を継承（JoinNode と共通の描画ロジック）
 */
export class ForkNode extends SyncBarNode {
	constructor(nodeData: ActivityNodeItem) {
		super(nodeData);
	}
}

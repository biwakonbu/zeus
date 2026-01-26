// ビューワー共通ユーティリティ関数
// Activity/UseCase 等のビューワー間で共有するヘルパー関数

/**
 * ガード条件を UML 準拠の表示形式にフォーマット
 * - 既に角括弧で囲まれている場合はそのまま返す
 * - 角括弧がない場合は追加する
 * - 空文字列や null/undefined は空文字列を返す
 *
 * @param guard - ガード条件文字列
 * @returns UML 準拠の `[条件]` 形式の文字列
 *
 * @example
 * formatGuardCondition('[条件]')  // => '[条件]'
 * formatGuardCondition('条件')    // => '[条件]'
 * formatGuardCondition('')        // => ''
 * formatGuardCondition(undefined) // => ''
 */
export function formatGuardCondition(guard: string | undefined | null): string {
	if (!guard) return '';

	// 既に角括弧で囲まれている場合はそのまま返す
	if (guard.startsWith('[') && guard.endsWith(']')) {
		return guard;
	}

	// 角括弧を追加して返す
	return `[${guard}]`;
}

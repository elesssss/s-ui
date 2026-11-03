// lib/rules/syntaxes/v-memo.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var v_memo_default = {
	supported: ">=3.2.0",
	createTemplateBodyVisitor(context) {
		/**
		* Reports `v-is` node
		*/
		function reportVMemo(vMemoAttr) {
			context.report({
				node: vMemoAttr.key,
				messageId: "forbiddenVMemo"
			});
		}
		return { "VAttribute[directive=true][key.name.name='memo']": reportVMemo };
	}
};
exports.default = v_memo_default;

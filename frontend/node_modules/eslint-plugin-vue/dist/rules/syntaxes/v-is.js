// lib/rules/syntaxes/v-is.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var v_is_default = {
	deprecated: "3.1.0",
	supported: ">=3.0.0",
	createTemplateBodyVisitor(context) {
		/**
		* Reports `v-is` node
		*/
		function reportVIs(vIsAttr) {
			context.report({
				node: vIsAttr.key,
				messageId: "forbiddenVIs"
			});
		}
		return { "VAttribute[directive=true][key.name.name='is']": reportVIs };
	}
};
exports.default = v_is_default;

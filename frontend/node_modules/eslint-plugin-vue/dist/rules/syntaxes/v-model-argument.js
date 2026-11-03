// lib/rules/syntaxes/v-model-argument.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var v_model_argument_default = {
	supported: ">=3.0.0",
	createTemplateBodyVisitor(context) {
		return { "VAttribute[directive=true] > VDirectiveKey[name.name='model'][argument!=null]"(node) {
			context.report({
				node: node.argument,
				messageId: "forbiddenVModelArgument"
			});
		} };
	}
};
exports.default = v_model_argument_default;

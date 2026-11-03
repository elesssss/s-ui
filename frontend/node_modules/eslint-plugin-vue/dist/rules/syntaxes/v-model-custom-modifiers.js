// lib/rules/syntaxes/v-model-custom-modifiers.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
const BUILTIN_MODIFIERS = /* @__PURE__ */ new Set([
	"lazy",
	"number",
	"trim"
]);
var v_model_custom_modifiers_default = {
	supported: ">=3.0.0",
	createTemplateBodyVisitor(context) {
		return { "VAttribute[directive=true] > VDirectiveKey[name.name='model'][modifiers.length>0]"(node) {
			for (const modifier of node.modifiers) if (!BUILTIN_MODIFIERS.has(modifier.name)) context.report({
				node: modifier,
				messageId: "forbiddenVModelCustomModifiers"
			});
		} };
	}
};
exports.default = v_model_custom_modifiers_default;

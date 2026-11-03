// lib/rules/syntaxes/v-bind-attr-modifier.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var v_bind_attr_modifier_default = {
	supported: ">=3.2.0",
	createTemplateBodyVisitor(context) {
		/**
		* Reports `v-bind.attr` node
		*/
		function report(mod) {
			context.report({
				node: mod,
				messageId: "forbiddenVBindAttrModifier"
			});
		}
		return { "VAttribute[directive=true][key.name.name='bind']"(node) {
			const attrMod = node.key.modifiers.find((m) => m.name === "attr");
			if (attrMod) report(attrMod);
		} };
	}
};
exports.default = v_bind_attr_modifier_default;

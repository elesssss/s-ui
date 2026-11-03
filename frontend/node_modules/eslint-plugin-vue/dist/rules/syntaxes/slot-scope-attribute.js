const require_can_convert_to_v_slot = require("./utils/can-convert-to-v-slot.js");
// lib/rules/syntaxes/slot-scope-attribute.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var slot_scope_attribute_default = {
	deprecated: "2.6.0",
	supported: ">=2.5.0 <3.0.0",
	createTemplateBodyVisitor(context, { fixToUpgrade } = {}) {
		const sourceCode = context.sourceCode;
		const tokenStore = sourceCode.parserServices.getTemplateBodyTokenStore && sourceCode.parserServices.getTemplateBodyTokenStore();
		/**
		* Checks whether the given node can convert to the `v-slot`.
		*/
		function canConvertToVSlot$1(startTag) {
			if (!require_can_convert_to_v_slot.default(startTag.parent, sourceCode, tokenStore)) return false;
			if (startTag.attributes.find((attr) => attr.directive === false && attr.key.name === "slot")) return false;
			return !startTag.attributes.find((attr) => attr.directive === true && attr.key.name.name === "bind" && attr.key.argument && attr.key.argument.type === "VIdentifier" && attr.key.argument.name === "slot");
		}
		/**
		* Convert to `v-slot`.
		*/
		function fixSlotScopeToVSlot(fixer, scopeAttr) {
			const element = scopeAttr.parent.parent;
			const replaceText = `v-slot${scopeAttr && scopeAttr.value ? `=${sourceCode.getText(scopeAttr.value)}` : ""}`;
			if (element.name === "template") return [fixer.replaceText(scopeAttr, replaceText)];
			const tokenBefore = tokenStore.getTokenBefore(scopeAttr);
			return [
				fixer.removeRange([tokenBefore.range[1], scopeAttr.range[1]]),
				fixer.insertTextBefore(element, `<template ${replaceText}>\n`),
				fixer.insertTextAfter(element, `\n</template>`)
			];
		}
		/**
		* Reports `slot-scope` node
		*/
		function reportSlotScope(scopeAttr) {
			context.report({
				node: scopeAttr.key,
				messageId: "forbiddenSlotScopeAttribute",
				fix(fixer) {
					if (!fixToUpgrade) return null;
					const startTag = scopeAttr.parent;
					return canConvertToVSlot$1(startTag) ? fixSlotScopeToVSlot(fixer, scopeAttr) : null;
				}
			});
		}
		return { "VAttribute[directive=true][key.name.name='slot-scope']": reportSlotScope };
	}
};
exports.default = slot_scope_attribute_default;

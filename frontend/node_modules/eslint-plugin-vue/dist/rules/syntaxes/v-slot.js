// lib/rules/syntaxes/v-slot.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
/**
* Checks whether the given node can convert to the `slot`.
*/
function canConvertToSlot(vSlotAttr) {
	return vSlotAttr.parent.parent.name === "template";
}
var v_slot_default = {
	supported: ">=2.6.0",
	createTemplateBodyVisitor(context) {
		const sourceCode = context.sourceCode;
		/**
		* Convert to `slot` and `slot-scope`.
		*/
		function fixVSlotToSlot(fixer, vSlotAttr) {
			const key = vSlotAttr.key;
			if (key.modifiers.length > 0) return null;
			const attrs = [];
			const argument = key.argument;
			if (argument) {
				if (argument.type === "VIdentifier") {
					const name = argument.rawName;
					attrs.push(`slot="${name}"`);
				} else if (argument.type === "VExpressionContainer" && argument.expression) {
					const expression = sourceCode.getText(argument.expression);
					attrs.push(`:slot="${expression}"`);
				} else return null;
			}
			const scopedValueNode = vSlotAttr.value;
			if (scopedValueNode) attrs.push(`slot-scope=${sourceCode.getText(scopedValueNode)}`);
			if (attrs.length === 0) attrs.push("slot");
			return fixer.replaceText(vSlotAttr, attrs.join(" "));
		}
		/**
		* Reports `v-slot` node
		*/
		function reportVSlot(vSlotAttr) {
			context.report({
				node: vSlotAttr.key,
				messageId: "forbiddenVSlot",
				fix(fixer) {
					return canConvertToSlot(vSlotAttr) ? fixVSlotToSlot(fixer, vSlotAttr) : null;
				}
			});
		}
		return { "VAttribute[directive=true][key.name.name='slot']": reportVSlot };
	}
};
exports.default = v_slot_default;

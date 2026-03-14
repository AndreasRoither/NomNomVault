export function formatEnumLabel(value?: string) {
	if (!value) {
		return undefined;
	}

	return value
		.split("_")
		.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
		.join(" ");
}

export function formatMinutes(value?: number) {
	if (!value || value <= 0) {
		return undefined;
	}

	return `${value} min`;
}

export function formatServings(value?: number) {
	if (!value || value <= 0) {
		return undefined;
	}

	return `${value} servings`;
}

export function formatCalories(value?: number) {
	if (!value || value <= 0) {
		return undefined;
	}

	return `${value} kcal`;
}

export function formatQuantity(value?: number, unit?: string) {
	if (value == null) {
		return undefined;
	}

	const normalized = Number.isInteger(value) ? String(value) : value.toFixed(1);
	return unit ? `${normalized} ${unit}` : normalized;
}

export function formatDateLabel(value?: string) {
	if (!value) {
		return undefined;
	}

	const date = new Date(value);
	if (Number.isNaN(date.getTime())) {
		return undefined;
	}

	return new Intl.DateTimeFormat("en", {
		month: "short",
		day: "numeric",
		year: "numeric",
	}).format(date);
}

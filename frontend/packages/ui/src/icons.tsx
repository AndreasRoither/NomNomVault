import type { JSX } from "solid-js";

import { cn } from "./utils";

export type IconProps = JSX.SvgSVGAttributes<SVGSVGElement> & {
	size?: "sm" | "md" | "lg";
};

function iconSizeClass(size: IconProps["size"]) {
	switch (size) {
		case "sm":
			return "h-[var(--nnv-icon-sm)] w-[var(--nnv-icon-sm)]";
		case "lg":
			return "h-[var(--nnv-icon-lg)] w-[var(--nnv-icon-lg)]";
		default:
			return "h-[var(--nnv-icon-md)] w-[var(--nnv-icon-md)]";
	}
}

function IconBase(
	props: IconProps & {
		viewBox?: string;
		children: JSX.Element;
	},
) {
	return (
		<svg
			viewBox={props.viewBox ?? "0 0 24 24"}
			fill="none"
			stroke="currentColor"
			stroke-width="1.85"
			stroke-linecap="round"
			stroke-linejoin="round"
			aria-hidden="true"
			class={cn("shrink-0", iconSizeClass(props.size), props.class)}
			{...props}
		>
			{props.children}
		</svg>
	);
}

export function SearchIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<circle cx="11" cy="11" r="7" />
			<path d="m20 20-3.5-3.5" />
		</IconBase>
	);
}

export function MenuIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="M4 7h16" />
			<path d="M4 12h16" />
			<path d="M4 17h16" />
		</IconBase>
	);
}

export function MoreIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<circle cx="5" cy="12" r="1.2" fill="currentColor" stroke="none" />
			<circle cx="12" cy="12" r="1.2" fill="currentColor" stroke="none" />
			<circle cx="19" cy="12" r="1.2" fill="currentColor" stroke="none" />
		</IconBase>
	);
}

export function ProfileIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<circle cx="12" cy="8" r="3.25" />
			<path d="M5 19c1.7-3 4.1-4.5 7-4.5s5.3 1.5 7 4.5" />
		</IconBase>
	);
}

export function SunIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<circle cx="12" cy="12" r="4" />
			<path d="M12 2.5v2" />
			<path d="M12 19.5v2" />
			<path d="m4.93 4.93 1.41 1.41" />
			<path d="m17.66 17.66 1.41 1.41" />
			<path d="M2.5 12h2" />
			<path d="M19.5 12h2" />
			<path d="m4.93 19.07 1.41-1.41" />
			<path d="m17.66 6.34 1.41-1.41" />
		</IconBase>
	);
}

export function MoonIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="M20 15.5A8.5 8.5 0 1 1 8.5 4 6.8 6.8 0 0 0 20 15.5Z" />
		</IconBase>
	);
}

export function FilterIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="M4 7h16" />
			<path d="M7 12h10" />
			<path d="M10 17h4" />
		</IconBase>
	);
}

export function ChevronLeftIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="m15 5-7 7 7 7" />
		</IconBase>
	);
}

export function ChevronRightIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="m9 5 7 7-7 7" />
		</IconBase>
	);
}

export function RecipesIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="M7 4.5h10l2 4.5-7 10L5 9l2-4.5Z" />
			<path d="M9.5 9h5" />
			<path d="M10.5 12h3" />
		</IconBase>
	);
}

export function InfoIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<circle cx="12" cy="12" r="9" />
			<path d="M12 10v6" />
			<circle cx="12" cy="7" r="1" fill="currentColor" stroke="none" />
		</IconBase>
	);
}

export function HouseholdIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="M4 11.5 12 5l8 6.5" />
			<path d="M6.5 10.5V19h11v-8.5" />
			<path d="M10 19v-5h4v5" />
		</IconBase>
	);
}

export function ClearIcon(props: IconProps) {
	return (
		<IconBase {...props}>
			<path d="m7 7 10 10" />
			<path d="m17 7-10 10" />
		</IconBase>
	);
}

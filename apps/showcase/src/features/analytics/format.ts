const integerFormatter = new Intl.NumberFormat("en-US");
const compactFormatter = new Intl.NumberFormat("en-US", {
  notation: "compact",
  maximumFractionDigits: 1,
});
const percentFormatter = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 1,
});
const scoreFormatter = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 3,
  maximumFractionDigits: 3,
});
const dateTickSpacing = 80;

interface DateTickScale {
  range(): unknown[];
}

export function formatInteger(value: number): string {
  return integerFormatter.format(value);
}

export function formatCompact(value: number): string {
  return compactFormatter.format(value);
}

export function formatPercent(value: number): string {
  return `${percentFormatter.format(value)}%`;
}

export function formatScore(value: number | null): string {
  return value === null ? "—" : scoreFormatter.format(value);
}

export function formatChartDate(value: Date | string): string {
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    timeZone: "UTC",
  }).format(new Date(value));
}

export function responsiveDateTicks(dates: Date[]): (scale: DateTickScale) => Date[] {
  return (scale) => {
    if (dates.length <= 2) return dates;

    const range = scale.range();
    const firstPosition = Number(range[0]);
    const lastPosition = Number(range.at(-1));

    if (!Number.isFinite(firstPosition) || !Number.isFinite(lastPosition)) {
      return [dates[0]!, dates.at(-1)!];
    }

    const availableWidth = Math.abs(lastPosition - firstPosition);
    const tickCount = Math.min(
      dates.length,
      Math.max(2, Math.floor(availableWidth / dateTickSpacing) + 1),
    );

    if (tickCount === dates.length) return dates;

    const lastDateIndex = dates.length - 1;
    return Array.from({ length: tickCount }, (_, index) => {
      const dateIndex = Math.round((index * lastDateIndex) / (tickCount - 1));
      return dates[dateIndex]!;
    });
  };
}

export function formatFullUtcDate(value: Date | string): string {
  return new Intl.DateTimeFormat("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(value));
}

export function formatUtcTimestamp(value: string): string {
  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    timeZone: "UTC",
    timeZoneName: "short",
  }).format(new Date(value));
}

export function formatUtcTime(value: string): string {
  return new Intl.DateTimeFormat("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
    timeZone: "UTC",
  }).format(new Date(value));
}

export function toUtcDate(value: string): Date {
  return new Date(`${value}T00:00:00Z`);
}

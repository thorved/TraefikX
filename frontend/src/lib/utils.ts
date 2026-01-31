import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDistanceToNow(date: Date, options?: { addSuffix?: boolean }): string {
  const now = new Date();
  const diffInMs = now.getTime() - new Date(date).getTime();
  const diffInSecs = Math.floor(diffInMs / 1000);
  const diffInMins = Math.floor(diffInSecs / 60);
  const diffInHours = Math.floor(diffInMins / 60);
  const diffInDays = Math.floor(diffInHours / 24);

  let result = "";

  if (diffInSecs < 60) {
    result = "less than a minute";
  } else if (diffInMins < 60) {
    result = `${diffInMins} minute${diffInMins === 1 ? "" : "s"}`;
  } else if (diffInHours < 24) {
    result = `${diffInHours} hour${diffInHours === 1 ? "" : "s"}`;
  } else if (diffInDays < 30) {
    result = `${diffInDays} day${diffInDays === 1 ? "" : "s"}`;
  } else {
    const diffInMonths = Math.floor(diffInDays / 30);
    result = `${diffInMonths} month${diffInMonths === 1 ? "" : "s"}`;
  }

  if (options?.addSuffix) {
    result += " ago";
  }

  return result;
}

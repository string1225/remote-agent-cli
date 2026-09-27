export interface SearchTextPart {
  text: string;
  match: boolean;
}

export function normalizeDiscussionSearch(query: string): string {
  return query.trim().toLocaleLowerCase();
}

export function textMatchesDiscussionSearch(
  text: string,
  query: string,
): boolean {
  const normalized = normalizeDiscussionSearch(query);
  return normalized.length > 0 && text.toLocaleLowerCase().includes(normalized);
}

export function splitDiscussionSearchMatches(
  text: string,
  query: string,
): SearchTextPart[] {
  const normalized = normalizeDiscussionSearch(query);
  if (!normalized) return [{ text, match: false }];

  const lowered = text.toLocaleLowerCase();
  const parts: SearchTextPart[] = [];
  let cursor = 0;
  while (cursor < text.length) {
    const matchIndex = lowered.indexOf(normalized, cursor);
    if (matchIndex === -1) {
      parts.push({ text: text.slice(cursor), match: false });
      break;
    }
    if (matchIndex > cursor)
      parts.push({ text: text.slice(cursor, matchIndex), match: false });
    parts.push({
      text: text.slice(matchIndex, matchIndex + normalized.length),
      match: true,
    });
    cursor = matchIndex + normalized.length;
  }
  return parts.length ? parts : [{ text, match: false }];
}

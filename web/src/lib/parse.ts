import { capitalize } from './dates';

export interface Entry {
  name: string;
  qty: string;
}

// "2 latte", "2x latte", "500 g farina", "1,5 kg patate". The mandatory space before the
// name keeps "2 latte" from reading as 2 l of "atte".
const LEADING_QUANTITY = /^(\d+(?:[.,]\d+)?)(?:\s?(kg|g|ml|l|pz))?(?:\s*x)?\s+(.+)$/i;
// "latte x2"
const TRAILING_QUANTITY = /^(.+?)\s*[x×]\s*(\d+)$/i;

export function parseEntry(raw: string): Entry | null {
  let name = raw.trim().replace(/\s+/g, ' ');
  let qty = '';

  const leading = name.match(LEADING_QUANTITY);
  const trailing = leading ? null : name.match(TRAILING_QUANTITY);
  if (leading) {
    const [, amount, unit, rest] = leading;
    qty = unit ? `${amount} ${unit.toLowerCase()}` : amount;
    name = rest;
  } else if (trailing) {
    [, name, qty] = trailing;
  }

  name = name.trim();
  return name ? { name: capitalize(name), qty } : null;
}

// A bare count reads as "×2"; a measure ("500 g") stays as written.
export function quantityLabel(qty: string): string {
  return /^\d+$/.test(qty) ? `×${qty}` : qty;
}

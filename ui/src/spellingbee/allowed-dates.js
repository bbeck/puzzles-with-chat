import isBefore from "date-fns/isBefore";
import isFuture from "date-fns/isFuture";

export const nytBeeFirstPuzzleDate = new Date(2018, 6, 29);
export function isNYTBeeDateAllowed(date) {
  return !isFuture(date) && !isBefore(date, nytBeeFirstPuzzleDate);
}

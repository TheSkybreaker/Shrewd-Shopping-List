import { clearCart, deleteItem, deleteList, toggleItem } from './actions';
import { isRedirecting } from './auth';
import { restoreList, saveList } from './cache';
import { fromIsoDate, fullDateLabel, monthLabel, relativeDay, weekdayLong } from './dates';
import { resumeDeletions } from './deletions';
import { listProgress, progressLabel, splitItems } from './groups';
import { goHome, goTo, setFlash } from './nav';
import type { Item } from './pb';
import { quantityLabel } from './parse';
import { loadList, loadUsers, onReconnect, subscribeList } from './realtime';
import { items, lists, session } from './stores';

const CHECK_ANIMATION_MS = 260;

const reducedMotion = () => matchMedia('(prefers-reduced-motion: reduce)').matches;

export function listPage() {
  const params = new URLSearchParams(location.search);

  return {
    listId: params.get('id') ?? '',
    focusComposer: params.has('nuova'),
    loaded: false,
    sheet: null as 'menu' | null,
    confirmDelete: false,
    // Items in the middle of the check animation, before they change section.
    checking: {} as Record<string, 'on' | 'off'>,

    async init() {
      if (isRedirecting()) return;
      if (!this.listId) return goTo('/');

      this.loaded = restoreList(this.listId);
      resumeDeletions();
      subscribeList(this.listId, () => {
        setFlash(`Lista eliminata da ${session().partner?.name ?? ''}`.trim());
        goTo('/');
      });
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') void this.reload();
        else saveList(this.listId);
      });
      onReconnect(() => void this.reload());
      await this.reload();
    },

    async reload() {
      try {
        const [, found] = await Promise.all([loadUsers(), loadList(this.listId)]);
        if (found) saveList(this.listId);
        else goTo('/');
      } catch {
        // Offline or a server error: the page keeps what it already shows.
      } finally {
        this.loaded = true;
      }
    },

    get list() {
      return lists().byId[this.listId];
    },

    get groups() {
      return splitItems(items().forList(this.listId));
    },

    get progress() {
      return listProgress(items().forList(this.listId));
    },

    get progressText() {
      return progressLabel(this.progress);
    },

    get day() {
      return this.list ? fromIsoDate(this.list.date).getDate() : '';
    },

    get relative() {
      return this.list ? relativeDay(this.list.date) : null;
    },

    get weekday() {
      return this.list ? weekdayLong(this.list.date) : '';
    },

    get month() {
      return this.list ? monthLabel(this.list.date) : '';
    },

    get dateLine() {
      return this.list ? fullDateLabel(this.list.date) : '';
    },

    quantity: quantityLabel,

    meta(item: Item) {
      const userId = item.checked ? item.checked_by || item.added_by : item.added_by;
      return {
        userId,
        text: `${item.checked ? 'preso da' : 'aggiunto da'} ${session().who(userId)}`,
      };
    },

    itemClass(item: Item) {
      return {
        'is-done': item.checked,
        'is-checking': this.checking[item.id] === 'on',
        'is-unchecking': this.checking[item.id] === 'off',
        'is-fresh': items().fresh[item.id] === true,
      };
    },

    toggle(item: Item) {
      if (this.checking[item.id]) return;
      if (reducedMotion()) return toggleItem(item.id);

      this.checking[item.id] = item.checked ? 'off' : 'on';
      setTimeout(() => {
        delete this.checking[item.id];
        toggleItem(item.id);
      }, CHECK_ANIMATION_MS);
    },

    remove(item: Item) {
      deleteItem(item);
    },

    openMenu() {
      this.confirmDelete = false;
      this.sheet = 'menu';
    },

    clearCart() {
      this.sheet = null;
      clearCart(this.listId);
    },

    deleteList() {
      if (!this.confirmDelete) {
        this.confirmDelete = true;
        return;
      }
      if (this.list) deleteList(this.list);
      goHome();
    },

    back: goHome,
  };
}

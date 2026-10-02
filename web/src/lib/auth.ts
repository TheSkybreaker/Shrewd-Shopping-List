import { ClientResponseError } from 'pocketbase';
import { clearCache } from './cache';
import { pb } from './pb';
import { forgetDevice } from './push';

export type PageName = 'login' | 'home' | 'list';

let redirecting = false;

// Pages skip their setup while the guard is sending the browser elsewhere.
export function isRedirecting(): boolean {
  return redirecting;
}

function redirect(url: string) {
  redirecting = true;
  location.replace(url);
}

export function guardPage(page: PageName) {
  const signedIn = pb.authStore.isValid;
  if (page === 'login' && signedIn) redirect('/');
  if (page !== 'login' && !signedIn) redirect('/login/');
}

// Renews the token and the user record. A network error keeps the session, so the app still
// opens offline; a rejected token signs out.
export async function refreshSession() {
  try {
    await pb.collection('users').authRefresh();
  } catch (error) {
    if (error instanceof ClientResponseError && [401, 403, 404].includes(error.status)) {
      pb.authStore.clear();
      redirect('/login/');
    }
  }
}

export async function logout() {
  await forgetDevice();
  clearCache();
  pb.authStore.clear();
  redirect('/login/');
}

export function loginPage() {
  return {
    email: '',
    password: '',
    error: '',
    busy: false,

    async submit() {
      this.busy = true;
      this.error = '';
      try {
        await pb.collection('users').authWithPassword(this.email.trim(), this.password);
        redirect('/');
      } catch {
        this.error = 'Email o password non corretti.';
        this.busy = false;
      }
    },
  };
}

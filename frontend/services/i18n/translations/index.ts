import type { Language } from '../types';
import ru from './ru';
import en from './en';

const translations: Record<Language, Record<string, string>> = { ru, en };

export default translations;

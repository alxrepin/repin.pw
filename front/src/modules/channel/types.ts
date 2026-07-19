export interface Channel {
  id: number;
  name: string;
  title: string;
  description?: string;
  avatar?: string;
  subscriptions: number;
  url: string;
}

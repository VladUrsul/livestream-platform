export interface Profile {
  user_id: string;
  username: string;
  email: string;
  display_name: string;
  bio: string;
  avatar_url: string;
  followers: number;
  following: number;
  is_live: boolean;
  created_at: string;
  updated_at: string;
}

export interface SearchResult {
  user_id: string;
  username: string;
  display_name: string;
  avatar_url: string;
  followers: number;
  is_live: boolean;
}
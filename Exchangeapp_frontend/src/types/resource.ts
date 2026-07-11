export interface ResourceSummary {
  id: number;
  authorId: number;
  title: string;
  preview: string;
  content: string;
  coverUrl?: string;
  contentImages?: string[];
  tags?: string[];
  status?: string;
  viewCount?: number;
  likeCount?: number;
  commentCount?: number;
  favoriteCount?: number;
  isFree?: boolean;
  requiredPoints?: number;
}

export interface ResourceDetail extends ResourceSummary {
  author?: {
    id: number;
    username: string;
  };
  stats?: {
    viewCount: number;
    likeCount: number;
    commentCount: number;
    favoriteCount: number;
  };
  isUnlocked?: boolean;
}

export interface ResourceLike {
  likes: number;
}

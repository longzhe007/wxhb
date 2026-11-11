"use client";

import { Video } from "lucide-react";

export function PromptCover({ coverUrl, title, className = "" }: { coverUrl?: string; title: string; className?: string }) {
    if (coverUrl) return <img src={coverUrl} alt={title} className={`aspect-[4/3] w-full object-cover ${className}`} />;
    return (
        <div className={`flex aspect-[4/3] w-full flex-col justify-between bg-stone-900 p-4 text-stone-100 dark:bg-stone-950 ${className}`}>
            <Video className="size-5 text-stone-400" aria-hidden />
            <div>
                <div className="text-[11px] text-stone-400">视频提示词</div>
                <div className="mt-1 line-clamp-2 text-sm font-medium leading-5">{title}</div>
            </div>
        </div>
    );
}

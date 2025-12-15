import { Component, Input, OnChanges, OnDestroy, SimpleChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { Thread, CharacterStats } from '../../services/character';

@Component({
selector: 'app-character-view',
standalone: true,
imports: [CommonModule, FormsModule],
templateUrl: './character-view.html',
styleUrl: './character-view.css',
})
export class CharacterViewComponent implements OnChanges, OnDestroy {
@Input() url!: string;
@Input() headerText: string = 'Character';
@Input() content: string = '';
@Input() loading: boolean = false;
@Input() error: string = '';
@Input() profileUrl?: string;
@Input() threads: Thread[] = [];
@Input() stats?: CharacterStats;
@Input() acSafe: boolean = false;
@Input() loadingString: string = 'Loading...';

// Local component state for toggling inactive threads
showInactive: boolean = false;

// Timer ID for loading rotation
private loadingTimer?: number;

constructor(
  private sanitizer: DomSanitizer
) {
}

ngOnChanges(changes: SimpleChanges) {
  // Check if loading state changed
  if (changes['loading']) {
    if (changes['loading'].currentValue === true) {
      // Started loading - start rotation
      this.startLoadingRotation();
    } else if (changes['loading'].currentValue === false) {
      // Stopped loading - clear timer
      this.stopLoadingRotation();
    }
  }
}

private startLoadingRotation() {
  this.stopLoadingRotation(); // Clear any existing timer
  this.rotateLoadingString();
}

private stopLoadingRotation() {
  if (this.loadingTimer !== undefined) {
    window.clearTimeout(this.loadingTimer);
    this.loadingTimer = undefined;
  }
}

rotateLoadingString() {
  const loadingStrings = [
    `Fetching your character's data...`,
    `Hold on, we're getting your character's data...`,
    `Just a sec, we're fetching your character's data...`,
    `Almost there, we're fetching your character's data...`,
    `We're almost done, we're fetching your character's data...`,
  ];

  if (this.loading) {
    this.loadingString = loadingStrings[Math.floor(Math.random() * loadingStrings.length)];
    this.loadingTimer = window.setTimeout(() => {
      this.rotateLoadingString();
    }, 1500);
  }
}

ngOnDestroy() {
  // Clean up timer when component is destroyed
  this.stopLoadingRotation();
}

get safeContent(): SafeHtml {
  const content = this.sanitizer.bypassSecurityTrustHtml(this.content);
  return content;
}

// Helper method to calculate days since last post
getDaysSincePost(date: Date): number {
  const now = new Date();
  const postDate = new Date(date);
  const diffTime = Math.abs(now.getTime() - postDate.getTime());
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  return diffDays;
}

// Helper method to get CSS classes for thread styling
getThreadClasses(thread: Thread): string[] {
  const classes: string[] = [];

  const daysSince = this.getDaysSincePost(thread.lastPostDate);

    // Active/closed status - supercedes all others
    if (!thread.isActive) {
      classes.push('closed-thread');
      return classes;
    }

  // Age-based styling
  if (thread.needsReply &&daysSince > 30) {
    classes.push('very-old-post');
  } else if (thread.needsReply && daysSince > 14) {
    classes.push('old-post');
  }

  // Reply status
  if (thread.needsReply) {
    classes.push('needs-reply');
  }

  return classes;
}

// Helper method to format date for display
formatDate(date: Date): string {
  const postDate = new Date(date);
  const diffDays = this.getDaysSincePost(date);

  if (diffDays === 0) {
    return 'Today';
  } else if (diffDays === 1) {
    return 'Yesterday';
  } else if (diffDays < 7) {
    return `${diffDays} days ago`;
  } else {
    return postDate.toLocaleDateString();
  }
}

}
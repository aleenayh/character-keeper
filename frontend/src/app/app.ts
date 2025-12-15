import { Component, OnInit, ChangeDetectorRef, NgZone, ViewEncapsulation } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CharacterViewComponent } from './components/character-view/character-view';
import { CharacterService, Thread, CharacterStats } from './services/character';

interface CharacterData {
  id: number;
  headerText: string;
url: string;
content: string;
loading: boolean;
error: string;
profileUrl?: string;
threads: Thread[];
stats?: CharacterStats;
acSafe: boolean;
}

@Component({
selector: 'app-root',
standalone: true,
imports: [CommonModule, FormsModule, CharacterViewComponent],
templateUrl: './app.html',
styleUrl: './app.css',
encapsulation: ViewEncapsulation.None
})
export class AppComponent implements OnInit {
title = 'Character Keeper';
headerText = '';
newUrl = '';
characters: CharacterData[] = [];
viewString = '';
statusMessage = '';
statusType: 'success' | 'error' | '' = '';
currentTheme: 'default' | 'beach-day' | 'nightfall' | 'sorbet' = 'default';
showManagementButtons: number | null = null;

    // Service injected automatically
    constructor(
      private characterService: CharacterService,
      private cdr: ChangeDetectorRef,
      private ngZone: NgZone
    ) {}

    // Lifecycle hook - like useEffect(() => {}, []) in React
    ngOnInit() {
      this.characters = this.loadStoredCharacters();
      this.characters.forEach((char: CharacterData) => { this.loadCharacter(char); });
      this.loadTheme();
    }

    private loadTheme() {
      const savedTheme = localStorage.getItem('theme') as 'default' | 'beach-day' | 'nightfall' | 'sorbet' | null;
      if (savedTheme) {
        this.currentTheme = savedTheme;
        this.applyTheme(savedTheme);
      }
    }

    private applyTheme(theme: 'default' | 'beach-day' | 'nightfall' | 'sorbet') {
      const root = document.documentElement;
      if (theme === 'default') {
        root.removeAttribute('data-theme');
      } else {
        root.setAttribute('data-theme', theme);
      }
    }

    changeTheme(theme: 'default' | 'beach-day' | 'nightfall' | 'sorbet') {
      this.currentTheme = theme;
      this.applyTheme(theme);
      localStorage.setItem('theme', theme);
    }

    private extractUidFromUrl(url: string): number | null {
      try {
        const urlObj = new URL(url);
        const uid = urlObj.searchParams.get('uid');
        return uid ? parseInt(uid, 10) : null;
      } catch {
        return null;
      }
    }

    addCharacter() {
      if (!this.newUrl.trim()) return;
  
  const id = this.extractUidFromUrl(this.newUrl);
  if (!id) {
    alert('Invalid URL: must contain a uid parameter');
    return;
  }
  
  // Check for duplicates by ID
  if (this.characters.find(c => c.id === id)) {
    alert('Character already added');
    return;
  }

  const character: CharacterData = {
    id,
    headerText: this.headerText,
    url: this.newUrl,
    content: '',
    loading: true,
    error: '',
    profileUrl: undefined,
    threads: [],
    stats: undefined,
    acSafe: false
  };

  this.characters.push(character);
  this.saveCharacters();
  this.loadCharacter(character);
  this.newUrl = '';
    }

    loadCharacter(character: CharacterData) {
      const index = this.characters.findIndex(c => c.id === character.id);
      if (index === -1) return;
    
      // Update to loading state
      this.characters = this.characters.map((char, i) => 
        i === index ? { ...char, loading: true, error: '' } : char
      );
    
      this.characterService.getCharacterData(character.url).subscribe({
        next: (response) => {
          this.ngZone.run(() => {
            // Create completely new array with parsed data
            this.characters = this.characters.map((char, i) =>
              i === index ? {
                ...char,
                content: response.content,
                profileUrl: response.profileUrl,
                threads: response.threads,
                stats: response.stats,
                acSafe: response.acSafe,
                loading: false
              } : char
            );
            this.cdr.detectChanges();
          });
        },
        error: (err) => {
          this.ngZone.run(() => {
            console.error('Error loading character', character.id, err.message);
            // Create completely new array
            this.characters = this.characters.map((char, i) =>
              i === index ? { ...char, error: err.message, loading: false } : char
            );
            this.cdr.detectChanges();
          });
        }
      });
    }


    removeCharacter(id: number) {
      this.characters = this.characters.filter(c => c.id !== id);
      this.saveCharacters();
    }

    private saveCharacters() {
      const data = this.characters.map(c => ({ id: c.id, url: c.url, headerText: c.headerText }));
      localStorage.setItem('characters', JSON.stringify(data));
    }
    
    private loadStoredCharacters() {
      const stored = localStorage.getItem('characters');
      if (!stored) return [];

      try {
        const data = JSON.parse(stored) as Array<{id: number, url: string, headerText?: string}>;
        return data.map((item) => ({
          id: item.id,
          url: item.url,
          headerText: item.headerText || 'Character',
          content: '',
          loading: false,
          error: '',
          profileUrl: undefined,
          threads: [],
          stats: undefined,
          acSafe: false
        }));
      } catch {
        return [];
      }
    }

    // TO-DO list functionality
    todoCollapsed = false;

    toggleTodo() {
      this.todoCollapsed = !this.todoCollapsed;
    }

    // Get all threads that need replies across all characters
    getTodoThreads(): Array<Thread & { characterName: string, characterId: number }> {
      const todoThreads: Array<Thread & { characterName: string, characterId: number }> = [];

      this.characters.forEach(char => {
        if (!char.loading && !char.error) {
          char.threads
            .filter(thread => thread.needsReply && thread.isActive)
            .forEach(thread => {
              todoThreads.push({
                ...thread,
                characterName: char.headerText,
                characterId: char.id
              });
            });
        }
      });

      // Sort by date - oldest first
      return todoThreads.sort((a, b) => {
        return new Date(a.lastPostDate).getTime() - new Date(b.lastPostDate).getTime();
      });
    }

    // Helper to calculate days since post
    getDaysSincePost(date: Date): number {
      const now = new Date();
      const postDate = new Date(date);
      const diffTime = Math.abs(now.getTime() - postDate.getTime());
      const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
      return diffDays;
    }

    // Helper to format date
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

    // Helper to get CSS classes for todo items
    getTodoItemClasses(thread: Thread): string[] {
      const classes: string[] = [];
      const daysSince = this.getDaysSincePost(thread.lastPostDate);

      if (daysSince > 30) {
        classes.push('very-old-post');
      } else if (daysSince > 14) {
        classes.push('old-post');
      }

      return classes;
    }

    showStatus(message: string, type: 'success' | 'error') {
      this.statusMessage = message;
      this.statusType = type;
      setTimeout(() => {
        this.statusMessage = '';
        this.statusType = '';
      }, 5000);
    }

    loadView() {
      const key = this.viewString.trim();
      if (!key) {
        this.showStatus('Please enter a view key', 'error');
        return;
      }

      this.characterService.loadCharacterData(key).subscribe({
        next: (response) => {
          this.ngZone.run(() => {
            const savedData = response.data;

            if (!savedData || !savedData.characters) {
              this.showStatus('Invalid save data format', 'error');
              return;
            }

            this.characters = savedData.characters.map((char: any) => ({
              id: char.id || char.uid,
              headerText: char.headerText || 'Character',
              url: char.url,
              content: char.content || '',
              loading: false,
              error: '',
              profileUrl: char.profileUrl,
              threads: char.threads || [],
              stats: char.stats
            }));

            this.saveCharacters();
            this.cdr.detectChanges();

            this.showStatus(`Successfully loaded "${key}" with ${this.characters.length} character(s)!`, 'success');
          });
        },
        error: (err) => {
          console.error('Load error:', err);
          if (err.status === 404) {
            this.showStatus(`Save key "${key}" not found. Check the key and try again.`, 'error');
          } else {
            this.showStatus(`Failed to load: ${err.error || err.message || 'Unknown error'}`, 'error');
          }
        }
      });
    }

    saveView() {
      const key = this.viewString.trim();
      if (!key) {
        this.showStatus('Please enter a view key', 'error');
        return;
      }

      if (this.characters.length === 0) {
        this.showStatus('No characters to save!', 'error');
        return;
      }

      const charactersToSave = this.characters.map(char => ({
        id: char.id,
        headerText: char.headerText,
        url: char.url,
        content: char.content,
        profileUrl: char.profileUrl,
        threads: char.threads,
        stats: char.stats
      }));

      this.characterService.saveCharacterData(key, charactersToSave).subscribe({
        next: (response) => {
          this.showStatus(`Successfully saved "${key}" with ${this.characters.length} character(s)!`, 'success');
        },
        error: (err) => {
          console.error('Save error:', err);
          this.showStatus(`Failed to save: ${err.error || err.message || 'Unknown error'}`, 'error');
        }
      });
    }

    reorderCharacters(uid:number, move: "up" | "down") {
      const index = this.characters.findIndex(c => c.id === uid);
      if (index === -1) return;
      const newIndex = move === "up" ? index - 1 : index + 1;
      if (newIndex < 0 || newIndex >= this.characters.length) return;
      const character = this.characters[index];
      this.characters = this.characters.map((char, i) =>
        i === index ? this.characters[newIndex] :
        i === newIndex ? character : char
      );
      this.characters[newIndex] = character;
      this.saveCharacters();
    }

    toggleManagementButtons(uid:number) {
					if (this.showManagementButtons === uid) {
						this.showManagementButtons = null;
					} else {
						this.showManagementButtons = uid;
					}
				}
}
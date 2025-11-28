import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, Observable, switchMap, of } from 'rxjs';
import { environment } from '../../environments/environment';

export interface Thread {
  id: string;
  title: string;
  url: string;
  participants: string;
  lastPoster: string;
  lastPostDate: Date;
  lastPosterUid: number | null;
  isActive: boolean;
  needsReply: boolean;
}

export interface CharacterStats {
  age: number | null;
  blood: string;
  ship: string;
  birthday: string | null;
  calculatedAge: number | null;
  ageNeedsUpdate: boolean;
}

export interface CharacterResponse {
  url: string;
  content: string;
  error?: string;
  profileUrl?: string;
  threads: Thread[];
  stats?: CharacterStats;
  }

  @Injectable({
    providedIn: 'root',
  })
export class CharacterService {

    private apiUrl = `${environment.apiUrl}/character`;
    private baseApiUrl = environment.apiUrl;

        // HttpClient is injected - like passing props but automatic
    constructor(private http: HttpClient) {}

    // Helper function to extract UID from profile link
    private extractUidFromProfileLink(link: string): number | null {
      try {
        const match = link.match(/uid=(\d+)/);
        return match ? parseInt(match[1], 10) : null;
      } catch {
        return null;
      }
    }

    // Helper function to parse date from string
    private parseThreadDate(dateStr: string): Date {
      // Assuming format like "Yesterday, 03:45 PM" or "01-15-2025, 10:30 AM"
      // This is a placeholder - actual implementation depends on the exact format
      try {
        // Remove extra whitespace
        const cleaned = dateStr.trim();

        // Handle "Yesterday"
        if (cleaned.toLowerCase().startsWith('yesterday')) {
          const yesterday = new Date();
          yesterday.setDate(yesterday.getDate() - 1);
          return yesterday;
        }

        // Handle "Today"
        if (cleaned.toLowerCase().startsWith('today')) {
          return new Date();
        }

        // Try to parse as regular date
        return new Date(cleaned);
      } catch {
        return new Date(); // Fallback to current date
      }
    }

    // Build profile link from character UID
    private buildProfileLink(characterUid: number): string | undefined {
      if (!characterUid || characterUid === 0) return undefined;
      return `https://charmingrp.com/member.php?action=profile&uid=${characterUid}`;
    }

    // Parse threads from threadlog table
    private parseThreads(doc: Document, characterUid: number): Thread[] {
      const threads: Thread[] = [];

      // Find all thread rows in the threadlog table
      // Assuming structure: <table id="threadlog"><tr class="active|closed">...</tr></table>
      const threadRows = doc.querySelectorAll('#threadlog tr.active, #threadlog tr.closed');

      threadRows.forEach((row, index) => {
        try {
          const isActive = row.classList.contains('active');

          // Extract thread title and URL (typically in first td > a)
          const titleLink = row.querySelector('td:first-child a') || row.querySelector('a[href*="showthread"]');
          const title = titleLink?.textContent?.trim() || 'Unknown Thread';
          const url = titleLink?.getAttribute('href') || '';

          // Extract participants (typically in a specific column)
          const participantsCells = row.querySelectorAll('td');
          let participants = '';
          if (participantsCells.length > 1) {
            // Usually participants are in the second column
            participants = participantsCells[1]?.textContent?.trim() || '';
          }

          // Extract last poster info (typically last column with "Last post by <a>...")
          const lastPostCell = row.querySelector('td:last-child');
          const lastPostLink = lastPostCell?.querySelector('a[href*="member.php"]');
          const lastPoster = lastPostLink?.textContent?.trim() || 'Unknown';
          const lastPosterUid = lastPostLink ? this.extractUidFromProfileLink(lastPostLink.getAttribute('href') || '') : null;

          // Extract date (typically text after "on" or similar)
          const dateText = lastPostCell?.textContent || '';
          const dateMatch = dateText.match(/on\s+(.+?)$/i) || dateText.match(/(\d{2}-\d{2}-\d{4}.*)/);
          const lastPostDate = dateMatch ? this.parseThreadDate(dateMatch[1]) : new Date();

          // Determine if thread needs reply
          const needsReply = lastPosterUid !== null && lastPosterUid !== characterUid;

          // Generate unique ID from URL or index
          const threadId = url.match(/tid=(\d+)/)?.[1] || `thread-${index}`;

          threads.push({
            id: threadId,
            title,
            url,
            participants,
            lastPoster,
            lastPostDate,
            lastPosterUid,
            isActive,
            needsReply
          });
        } catch (error) {
          console.warn('Failed to parse thread row:', error);
        }
      });

      return threads;
    }

    // Helper to calculate age based on fictional date (130 years ago)
    private calculateFictionalAge(birthday: string): number | null {
      try {
        const birthDate = new Date(birthday);
        if (Number.isNaN(birthDate.getTime())) return null;

        // Today's date in the fictional setting (130 years ago)
        const today = new Date();
        const fictionalYear = today.getFullYear() - 130;
        const fictionalToday = new Date(fictionalYear, today.getMonth(), today.getDate());

        let age = fictionalToday.getFullYear() - birthDate.getFullYear();
        const monthDiff = fictionalToday.getMonth() - birthDate.getMonth();
        const dayDiff = fictionalToday.getDate() - birthDate.getDate();

        // Adjust if birthday hasn't occurred yet this year
        if (monthDiff < 0 || (monthDiff === 0 && dayDiff < 0)) {
          age--;
        }

        return age;
      } catch {
        return null;
      }
    }

    // Parse character stats from profile page
    private parseCharacterStats(doc: Document): CharacterStats {
      const stats: CharacterStats = {
        age: null,
        blood: '',
        ship: '',
        birthday: null,
        calculatedAge: null,
        ageNeedsUpdate: false
      };

      try {
        // Parse age from mp-age element
        const ageElement = doc.querySelector('.mp-age');
        if (ageElement) {
          const ageText = ageElement.textContent?.trim() || '';
          const ageNum = parseInt(ageText, 10);
          if (!Number.isNaN(ageNum)) {
            stats.age = ageNum;
          }
        }

        // Parse blood type from mp-blood element
        const bloodElement = doc.querySelector('.mp-blood');
        if (bloodElement) {
          stats.blood = bloodElement.textContent?.trim() || '';
        }

        // Parse ship from mp-ship element
        const shipElement = doc.querySelector('.mp-ship');
        if (shipElement) {
          stats.ship = shipElement.textContent?.trim() || '';
        }

        // Parse birthday from mp-profilecontent > mp-scrollpad
        const profileContent = doc.querySelector('.mp-profilecontent .mp-scrollpad');
        if (profileContent) {
          const html = profileContent.innerHTML;
          const birthdayMatch = html.match(/<strong>Birthdate:<\/strong>\s*([^<]+)/i);
          if (birthdayMatch) {
            const birthdayStr = birthdayMatch[1].trim();
            stats.birthday = birthdayStr;

            // Calculate expected age
            stats.calculatedAge = this.calculateFictionalAge(birthdayStr);

            // Check if age needs update
            if (stats.age !== null && stats.calculatedAge !== null) {
              stats.ageNeedsUpdate = stats.age !== stats.calculatedAge;
            }
          }
        }
      } catch (error) {
        console.warn('Failed to parse character stats:', error);
      }

      return stats;
    }

    // Fetch and parse profile data
    private fetchProfileStats(profileUrl: string): Observable<CharacterStats> {
      console.log('Fetching profile stats from:', profileUrl);
      return this.http.get<CharacterResponse>(`${this.apiUrl}?url=${encodeURIComponent(profileUrl)}`).pipe(
        map(response => {
          const parser = new DOMParser();
          const doc = parser.parseFromString(response.content, 'text/html');
          const stats = this.parseCharacterStats(doc);
          console.log('Parsed stats:', stats);
          return stats;
        })
      );
    }

        // Returns Observable instead of Promise
    getCharacterData(url: string): Observable<CharacterResponse> {
      // Extract character UID from URL
      const uidMatch = url.match(/uid=(\d+)/);
      const characterUid = uidMatch ? parseInt(uidMatch[1], 10) : 0;

      return this.http.get<CharacterResponse>(`${this.apiUrl}?url=${encodeURIComponent(url)}`).pipe(
        switchMap(response => {
          // Parse HTML and extract threadlog
          const parser = new DOMParser();
          const doc = parser.parseFromString(response.content, 'text/html');
          const threadlog = doc.querySelector('#threadlog');
          const content = threadlog ? threadlog.outerHTML : 'Thread log not found';

          // Build profile URL from character UID
          const profileUrl = this.buildProfileLink(characterUid);
          const threads = this.parseThreads(doc, characterUid);

          // If we have a profile URL, fetch the stats
          if (profileUrl) {
            return this.fetchProfileStats(profileUrl).pipe(
              map(stats => ({
                url,
                content,
                profileUrl,
                threads,
                stats
              }))
            );
          } else {
            // No profile URL, return without stats
            return of({
              url,
              content,
              profileUrl,
              threads
            });
          }
        })
      );
    }

    saveCharacterData(key: string, characters: any[]): Observable<any> {
      return this.http.post(`${this.baseApiUrl}/save`, {
        key: key,
        data: {
          savedAt: new Date().toISOString(),
          characters: characters
        }
      });
    }

    loadCharacterData(key: string): Observable<any> {
      return this.http.get(`${this.baseApiUrl}/load?key=${encodeURIComponent(key)}`);
    }

    deleteCharacterData(key: string): Observable<any> {
      return this.http.delete(`${this.baseApiUrl}/delete?key=${encodeURIComponent(key)}`);
    }
}
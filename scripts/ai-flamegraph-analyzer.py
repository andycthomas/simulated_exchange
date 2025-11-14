#!/usr/bin/env python3
"""
AI-Powered Flamegraph Analyzer

This script analyzes Go pprof profiles and generates actionable insights
using pattern matching and heuristics.

Usage: python3 ai-flamegraph-analyzer.py <profile.pb.gz>
"""

import sys
import subprocess
import re
import os
from datetime import datetime
from pathlib import Path


class FlamegraphAnalyzer:
    def __init__(self, profile_path):
        self.profile_path = profile_path
        self.profile_name = Path(profile_path).stem
        self.analysis = []
        self.hotspots = []
        self.recommendations = []

    def run_pprof_top(self, nodecount=50):
        """Run pprof -top to get function statistics"""
        try:
            result = subprocess.run(
                ['go', 'tool', 'pprof', '-top', f'-nodecount={nodecount}', self.profile_path],
                capture_output=True,
                text=True,
                timeout=30
            )
            return result.stdout
        except Exception as e:
            print(f"Error running pprof: {e}", file=sys.stderr)
            return ""

    def parse_top_output(self, top_output):
        """Parse pprof -top output into structured data"""
        functions = []
        lines = top_output.split('\n')

        # Skip header lines
        data_started = False
        for line in lines:
            if 'flat' in line and 'cum' in line:
                data_started = True
                continue

            if not data_started or not line.strip():
                continue

            # Parse line format: flat  flat%   sum%        cum   cum%
            parts = line.split()
            if len(parts) >= 6:
                try:
                    flat_time = parts[0]
                    flat_pct = parts[1].rstrip('%')
                    cum_time = parts[3]
                    cum_pct = parts[4].rstrip('%')
                    func_name = ' '.join(parts[5:])

                    functions.append({
                        'flat_time': flat_time,
                        'flat_pct': float(flat_pct),
                        'cum_time': cum_time,
                        'cum_pct': float(cum_pct),
                        'name': func_name
                    })
                except (ValueError, IndexError):
                    continue

        return functions

    def identify_hotspots(self, functions):
        """Identify performance hotspots"""
        hotspots = []

        for func in functions:
            priority = 'LOW'
            issues = []

            # High flat percentage = direct CPU consumption
            if func['flat_pct'] > 10:
                priority = 'HIGH'
                issues.append(f"Consumes {func['flat_pct']:.1f}% CPU directly")
            elif func['flat_pct'] > 5:
                priority = 'MEDIUM'
                issues.append(f"Consumes {func['flat_pct']:.1f}% CPU directly")

            # High cumulative percentage = hot call path
            if func['cum_pct'] > 20:
                if priority != 'HIGH':
                    priority = 'MEDIUM'
                issues.append(f"Call tree consumes {func['cum_pct']:.1f}% total CPU")

            # Pattern matching for common issues
            if any(pattern in func['name'].lower() for pattern in ['gc', 'garbage', 'sweep']):
                issues.append("Garbage collection overhead")
                if func['flat_pct'] > 5:
                    priority = 'HIGH'

            if any(pattern in func['name'].lower() for pattern in ['lock', 'mutex', 'sync']):
                issues.append("Synchronization/locking detected")
                if func['flat_pct'] > 3:
                    priority = 'HIGH'

            if any(pattern in func['name'].lower() for pattern in ['json', 'marshal', 'unmarshal']):
                issues.append("JSON serialization overhead")
                if func['flat_pct'] > 5:
                    priority = 'MEDIUM'

            if any(pattern in func['name'].lower() for pattern in ['sql', 'query', 'exec']):
                issues.append("Database query overhead")
                if func['flat_pct'] > 5:
                    priority = 'HIGH'

            if 'reflect' in func['name'].lower():
                issues.append("Reflection usage (slow)")
                if func['flat_pct'] > 3:
                    priority = 'MEDIUM'

            if any(pattern in func['name'].lower() for pattern in ['regex', 'regexp']):
                issues.append("Regular expression overhead")

            if 'syscall' in func['name'].lower():
                issues.append("System call overhead")

            if issues:
                hotspots.append({
                    'function': func['name'],
                    'flat_pct': func['flat_pct'],
                    'cum_pct': func['cum_pct'],
                    'priority': priority,
                    'issues': issues
                })

        return hotspots

    def generate_recommendations(self, hotspots):
        """Generate actionable recommendations"""
        recommendations = []

        # Group by issue type
        issue_groups = {
            'gc': [],
            'lock': [],
            'json': [],
            'sql': [],
            'reflect': [],
            'regex': [],
            'syscall': []
        }

        for hotspot in hotspots:
            for issue in hotspot['issues']:
                if 'garbage collection' in issue.lower():
                    issue_groups['gc'].append(hotspot)
                elif 'synchronization' in issue.lower() or 'locking' in issue.lower():
                    issue_groups['lock'].append(hotspot)
                elif 'json' in issue.lower():
                    issue_groups['json'].append(hotspot)
                elif 'database' in issue.lower():
                    issue_groups['sql'].append(hotspot)
                elif 'reflection' in issue.lower():
                    issue_groups['reflect'].append(hotspot)
                elif 'regular expression' in issue.lower():
                    issue_groups['regex'].append(hotspot)
                elif 'system call' in issue.lower():
                    issue_groups['syscall'].append(hotspot)

        # Generate recommendations
        if issue_groups['gc']:
            total_gc = sum(h['flat_pct'] for h in issue_groups['gc'])
            recommendations.append({
                'category': 'Memory Management',
                'priority': 'HIGH' if total_gc > 10 else 'MEDIUM',
                'issue': f'GC overhead: {total_gc:.1f}% CPU',
                'suggestions': [
                    'Reduce allocations using object pooling (sync.Pool)',
                    'Pre-allocate slices and maps with known capacity',
                    'Use value types instead of pointers where possible',
                    'Consider increasing GOGC value if memory permits'
                ]
            })

        if issue_groups['lock']:
            total_lock = sum(h['flat_pct'] for h in issue_groups['lock'])
            recommendations.append({
                'category': 'Concurrency',
                'priority': 'HIGH' if total_lock > 5 else 'MEDIUM',
                'issue': f'Lock contention: {total_lock:.1f}% CPU',
                'suggestions': [
                    'Reduce lock granularity (use more fine-grained locks)',
                    'Replace mutexes with channels where appropriate',
                    'Use sync.RWMutex for read-heavy workloads',
                    'Consider lock-free data structures',
                    'Run blocking profile: make profile-block'
                ]
            })

        if issue_groups['json']:
            total_json = sum(h['flat_pct'] for h in issue_groups['json'])
            recommendations.append({
                'category': 'Serialization',
                'priority': 'MEDIUM' if total_json > 5 else 'LOW',
                'issue': f'JSON overhead: {total_json:.1f}% CPU',
                'suggestions': [
                    'Use json.RawMessage for delayed parsing',
                    'Consider faster alternatives (easyjson, jsoniter)',
                    'Batch JSON operations',
                    'Use streaming for large payloads'
                ]
            })

        if issue_groups['sql']:
            total_sql = sum(h['flat_pct'] for h in issue_groups['sql'])
            recommendations.append({
                'category': 'Database',
                'priority': 'HIGH' if total_sql > 10 else 'MEDIUM',
                'issue': f'Database overhead: {total_sql:.1f}% CPU',
                'suggestions': [
                    'Add indexes for slow queries',
                    'Use prepared statements',
                    'Implement caching layer (Redis)',
                    'Batch database operations',
                    'Consider connection pooling optimization'
                ]
            })

        if issue_groups['reflect']:
            total_reflect = sum(h['flat_pct'] for h in issue_groups['reflect'])
            recommendations.append({
                'category': 'Code Generation',
                'priority': 'MEDIUM',
                'issue': f'Reflection overhead: {total_reflect:.1f}% CPU',
                'suggestions': [
                    'Replace reflection with code generation',
                    'Cache reflection results',
                    'Use concrete types instead of interface{}',
                    'Consider compile-time alternatives'
                ]
            })

        return recommendations

    def generate_report(self):
        """Generate markdown report"""
        output_path = Path(self.profile_path).parent / f"{self.profile_name}_analysis.md"

        with open(output_path, 'w') as f:
            f.write(f"# Flamegraph Analysis Report\n\n")
            f.write(f"**Profile:** `{self.profile_name}`\n")
            f.write(f"**Generated:** {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
            f.write(f"**Profile Path:** `{self.profile_path}`\n\n")

            f.write("---\n\n")

            # Executive Summary
            f.write("## 🎯 Executive Summary\n\n")
            if self.hotspots:
                top_hotspot = self.hotspots[0]
                f.write(f"**Top Hotspot:** `{top_hotspot['function']}`\n")
                f.write(f"- **CPU Usage:** {top_hotspot['flat_pct']:.1f}% (direct)\n")
                f.write(f"- **Priority:** {top_hotspot['priority']}\n\n")

                high_priority = [h for h in self.hotspots if h['priority'] == 'HIGH']
                f.write(f"**Critical Issues:** {len(high_priority)} high-priority hotspots found\n\n")
            else:
                f.write("No significant hotspots detected. Performance appears optimal.\n\n")

            # Top Hotspots
            f.write("## 🔥 Top Performance Hotspots\n\n")
            if self.hotspots:
                for i, hotspot in enumerate(self.hotspots[:10], 1):
                    priority_emoji = {
                        'HIGH': '🔴',
                        'MEDIUM': '🟡',
                        'LOW': '🟢'
                    }[hotspot['priority']]

                    f.write(f"### {i}. {priority_emoji} {hotspot['priority']} Priority\n\n")
                    f.write(f"**Function:** `{hotspot['function']}`\n\n")
                    f.write(f"**Metrics:**\n")
                    f.write(f"- Direct CPU: {hotspot['flat_pct']:.1f}%\n")
                    f.write(f"- Total CPU (with callees): {hotspot['cum_pct']:.1f}%\n\n")

                    f.write(f"**Issues Identified:**\n")
                    for issue in hotspot['issues']:
                        f.write(f"- {issue}\n")
                    f.write("\n")
            else:
                f.write("No hotspots detected.\n\n")

            # Recommendations
            f.write("## 💡 Optimization Recommendations\n\n")
            if self.recommendations:
                for i, rec in enumerate(self.recommendations, 1):
                    priority_emoji = {
                        'HIGH': '🔴',
                        'MEDIUM': '🟡',
                        'LOW': '🟢'
                    }[rec['priority']]

                    f.write(f"### {i}. {priority_emoji} {rec['category']}\n\n")
                    f.write(f"**Priority:** {rec['priority']}\n")
                    f.write(f"**Issue:** {rec['issue']}\n\n")
                    f.write(f"**Suggested Actions:**\n")
                    for suggestion in rec['suggestions']:
                        f.write(f"- {suggestion}\n")
                    f.write("\n")
            else:
                f.write("No specific recommendations. Profile looks healthy!\n\n")

            # Next Steps
            f.write("## 🚀 Next Steps\n\n")
            f.write("1. **Review the flamegraph visually:**\n")
            f.write(f"   ```bash\n")
            f.write(f"   firefox {self.profile_path.replace('.pb.gz', '.svg')}\n")
            f.write(f"   ```\n\n")
            f.write("2. **Interactive analysis:**\n")
            f.write(f"   ```bash\n")
            f.write(f"   go tool pprof {self.profile_path}\n")
            f.write(f"   # Commands: top, list <function>, web\n")
            f.write(f"   ```\n\n")
            f.write("3. **Generate additional profiles:**\n")
            f.write("   ```bash\n")
            f.write("   make profile-heap      # Memory allocation profile\n")
            f.write("   make profile-goroutine # Goroutine usage\n")
            f.write("   make profile-block     # Blocking operations\n")
            f.write("   make profile-mutex     # Lock contention\n")
            f.write("   ```\n\n")

            # Footer
            f.write("---\n\n")
            f.write("*Generated by AI Flamegraph Analyzer*\n")

        return output_path

    def analyze(self):
        """Run full analysis"""
        print(f"🔍 Analyzing profile: {self.profile_path}")
        print()

        # Get top functions
        top_output = self.run_pprof_top()
        if not top_output:
            print("❌ Failed to get profile data", file=sys.stderr)
            return None

        # Parse functions
        functions = self.parse_top_output(top_output)
        print(f"✓ Parsed {len(functions)} functions")

        # Identify hotspots
        self.hotspots = self.identify_hotspots(functions)
        print(f"✓ Identified {len(self.hotspots)} hotspots")

        # Generate recommendations
        self.recommendations = self.generate_recommendations(self.hotspots)
        print(f"✓ Generated {len(self.recommendations)} recommendations")
        print()

        # Generate report
        report_path = self.generate_report()
        print(f"📄 Analysis report generated: {report_path}")
        print()

        # Show quick summary
        if self.hotspots:
            print("🔥 Top 3 Hotspots:")
            for i, hotspot in enumerate(self.hotspots[:3], 1):
                priority_emoji = {'HIGH': '🔴', 'MEDIUM': '🟡', 'LOW': '🟢'}[hotspot['priority']]
                func_short = hotspot['function'][:60] + '...' if len(hotspot['function']) > 60 else hotspot['function']
                print(f"  {i}. {priority_emoji} {func_short}")
                print(f"     {hotspot['flat_pct']:.1f}% CPU - {hotspot['issues'][0]}")
            print()

        print(f"📖 Full report: cat {report_path}")
        print()

        return report_path


def main():
    if len(sys.argv) < 2:
        print("Usage: python3 ai-flamegraph-analyzer.py <profile.pb.gz>", file=sys.stderr)
        sys.exit(1)

    profile_path = sys.argv[1]

    if not os.path.exists(profile_path):
        print(f"Error: Profile not found: {profile_path}", file=sys.stderr)
        sys.exit(1)

    analyzer = FlamegraphAnalyzer(profile_path)
    report_path = analyzer.analyze()

    if report_path:
        sys.exit(0)
    else:
        sys.exit(1)


if __name__ == '__main__':
    main()

# 下游补丁序列

`series/` 里的补丁由 `git format-patch` 生成，基线由 `patches/base` 一行指定（当前为上游 tag `v2.7.0`，`github.com/go-vikunja/vikunja`）。
上游发新版时，CI（`.github/workflows/upstream-watch.yml`）会：取上游新 tag 的源码 → `git am -3 series/*.patch` → 跑测试 → 构建镜像。

**规则**

1. 改补丁 = 改源码 + **同步重新生成 `series/`**，两者必须一致。`upstream-watch.yml` 有个守卫：
   **main 的 `pkg/` 必须等于「`patches/base` 那个 tag + `series/`」**，否则拒绝构建并报错。
   （守卫比的是补丁基线，不是当天要构建的上游 tag —— 上游升版时 main 的 `pkg/` 本来就停在旧基线上，拿新 tag 比会误判成漂移。）
2. 重新生成（在本仓库根目录）：
   ```bash
   git fetch https://github.com/go-vikunja/vikunja.git 'refs/tags/v2.7.0:refs/tags/upstream-base'
   # 以 upstream-base 为基线重放你的改动，分成若干小 commit，然后：
   git format-patch --no-signature upstream-base..<你的分支> -o patches/series
   ```
3. 补丁只碰 `pkg/`（后端）。前端、Dockerfile、`.github/` 一律不改 —— 构建时用的是上游那份。
4. 上游把问题自己修了（或改了 caldav 导出）时，`git am -3` 会冲突：watch 会失败并按失败处理，不要去改 CI 去绕过，
   而是判断「补丁退役换官方镜像」还是「补丁改造」。
5. **换基线**（补丁基于新的上游 tag 重生成）时，除了重生成 `series/`，还要把 `patches/base` 改成新 tag，
   并让 main 的 `pkg/` 同步到「新 tag + 新补丁」，否则守卫会（正确地）拒绝构建。

**产物命名**：`<上游 tag>-tz.<n>`，例如 `v2.7.0-tz.2`。`n=1` 由 watch 自动构建；同一上游版本再出一版由人工 dispatch 指定。
浮动别名 `tz` 只在人工推 `v*` tag（`promote.yml`）时才移动 —— 那就是上线闸。

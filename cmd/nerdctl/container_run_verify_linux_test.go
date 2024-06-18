/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"
	"testing"

	"github.com/containerd/nerdctl/v2/pkg/testutil"
	"github.com/containerd/nerdctl/v2/pkg/testutil/testregistry"
)

func TestRunVerifyCosign(t *testing.T) {
	testutil.RequireExecutable(t, "cosign")
	testutil.DockerIncompatible(t)
	testutil.RequiresBuild(t)
	t.Parallel()

	base := testutil.NewBase(t)
	base.Env = append(base.Env, "COSIGN_PASSWORD=1")

	keyPair := newCosignKeyPair(t, "cosign-key-pair", "1")
	reg := testregistry.NewWithNoAuth(base, 0, false)
	t.Cleanup(func() {
		keyPair.cleanup()
		base.Cmd("builder", "prune").Run()
		reg.Cleanup(nil)
	})

	tID := testutil.Identifier(t)
	testImageRef := fmt.Sprintf("127.0.0.1:%d/%s", reg.Port, tID)
	dockerfile := fmt.Sprintf(`FROM %s
CMD ["echo", "nerdctl-build-test-string"]
	`, testutil.CommonImage)

	buildCtx := createBuildContext(t, dockerfile)

	base.Cmd("build", "-t", testImageRef, buildCtx).AssertOK()
	base.Cmd("push", testImageRef, "--sign=cosign", "--cosign-key="+keyPair.privateKey).AssertOK()
	base.Cmd("run", "--rm", "--verify=cosign", "--cosign-key="+keyPair.publicKey, testImageRef).AssertOK()
	base.Cmd("run", "--rm", "--verify=cosign", "--cosign-key=dummy", testImageRef).AssertFail()
}

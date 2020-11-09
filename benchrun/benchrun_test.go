/*
 * Copyright (c) 2020-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchrun

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestBenchRun(t *testing.T) {
	// run: go.exe test github.com/untillpro/dynobuffers/benchrun -run ^TestBenchRun$ -v
	cmd := exec.Command("go", "test", "github.com/untillpro/dynobuffers/benchmarks", "-bench", "^(Benchmark)", "-v", "-benchmem", "-race")

	var err error
	cmd.Dir, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Println(t)
	}

	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
}

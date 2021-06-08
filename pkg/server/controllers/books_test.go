/* Copyright (C) 2019, 2020 Monomax Software Pty Ltd
 *
 * This file is part of Dnote.
 *
 * Dnote is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Dnote is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Dnote.  If not, see <https://www.gnu.org/licenses/>.
 */

package controllers

import (
	"encoding/json"
	// "fmt"
	"net/http"
	// "net/url"
	"testing"
	// "time"

	"github.com/dnote/dnote/pkg/assert"
	"github.com/dnote/dnote/pkg/clock"
	"github.com/dnote/dnote/pkg/server/app"
	"github.com/dnote/dnote/pkg/server/config"
	"github.com/dnote/dnote/pkg/server/database"
	"github.com/dnote/dnote/pkg/server/presenters"
	"github.com/dnote/dnote/pkg/server/testutils"
	"github.com/pkg/errors"
)

func TestGetBooks(t *testing.T) {
	testutils.RunForWebAndAPI(t, "get notes", func(t *testing.T, target testutils.EndpointType) {
		defer testutils.ClearData(testutils.DB)

		// Setup
		server := MustNewServer(t, &app.App{
			Clock: clock.NewMock(),
			Config: config.Config{
				PageTemplateDir: "../views",
			},
		})
		defer server.Close()

		user := testutils.SetupUserData()
		anotherUser := testutils.SetupUserData()

		b1 := database.Book{
			UserID:  user.ID,
			Label:   "js",
			USN:     1123,
			Deleted: false,
		}
		testutils.MustExec(t, testutils.DB.Save(&b1), "preparing b1")
		b2 := database.Book{
			UserID:  user.ID,
			Label:   "css",
			USN:     1125,
			Deleted: false,
		}
		testutils.MustExec(t, testutils.DB.Save(&b2), "preparing b2")
		b3 := database.Book{
			UserID:  anotherUser.ID,
			Label:   "css",
			USN:     1128,
			Deleted: false,
		}
		testutils.MustExec(t, testutils.DB.Save(&b3), "preparing b3")
		b4 := database.Book{
			UserID:  user.ID,
			Label:   "",
			USN:     1129,
			Deleted: true,
		}
		testutils.MustExec(t, testutils.DB.Save(&b4), "preparing b4")

		// Execute
		var endpoint string
		if target == testutils.EndpointWeb {
			endpoint = "/books"
		} else {
			endpoint = "/api/v3/books"
		}

		req := testutils.MakeReq(server.URL, "GET", endpoint, "")
		res := testutils.HTTPAuthDo(t, req, user)

		// Test
		assert.StatusCodeEquals(t, res, http.StatusOK, "")

		if target == testutils.EndpointAPI {
			var payload []presenters.Book
			if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
				t.Fatal(errors.Wrap(err, "decoding payload"))
			}

			var b1Record, b2Record database.Book
			testutils.MustExec(t, testutils.DB.Where("id = ?", b1.ID).First(&b1Record), "finding b1")
			testutils.MustExec(t, testutils.DB.Where("id = ?", b2.ID).First(&b2Record), "finding b2")
			testutils.MustExec(t, testutils.DB.Where("id = ?", b2.ID).First(&b2Record), "finding b2")

			expected := []presenters.Book{
				{
					UUID:      b2Record.UUID,
					CreatedAt: b2Record.CreatedAt,
					UpdatedAt: b2Record.UpdatedAt,
					Label:     b2Record.Label,
					USN:       b2Record.USN,
				},
				{
					UUID:      b1Record.UUID,
					CreatedAt: b1Record.CreatedAt,
					UpdatedAt: b1Record.UpdatedAt,
					Label:     b1Record.Label,
					USN:       b1Record.USN,
				},
			}

			assert.DeepEqual(t, payload, expected, "payload mismatch")
		}
	})
}